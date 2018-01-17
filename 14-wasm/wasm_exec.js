const isNode = typeof process !== "undefined";
let args = ["js"];
let trace = false;
if (isNode) {
  args = args.concat(process.argv.slice(3));
  trace = process.env.TRACE === "1";
}

let nodeFS, nodeCrypto, encoder, decoder, mod;
let mem = new DataView(new ArrayBuffer(0)); // for editor autocompletion
let exitCode = 1;
let names = [];
let prevSP = 0;
let startSec = 0;
let outputBuf = "";
let dirContents = {};

const ENOSYS = 38;

if (isNode) {
  nodeFS = require("fs");
  nodeCrypto = require("crypto");
  const { TextEncoder, TextDecoder } = require("util");
  encoder = new TextEncoder("utf-8");
  decoder = new TextDecoder("utf-8");

  if (process.argv.length < 3) {
    process.stderr.write("usage: go_js_wasm_exec [wasm binary]\n");
    process.exit(1);
  }

  startSec = process.hrtime()[0];
  compileAndRun(nodeFS.readFileSync(process.argv[2])).catch((err) => {
    console.error(err);
  });
} else {
  encoder = new TextEncoder("utf-8");
  decoder = new TextDecoder("utf-8");
}

function nanotime() {
  let nanotime;
  if (isNode) {
    const [sec, nsec] = process.hrtime();
    nanotime = (sec - startSec) * 1000000000 + nsec;
  } else {
    nanotime = window.performance.now() * 1000000;
  }
  return nanotime;
}

async function compileAndRun(source) {
  await compile(source);
  await run();
}

async function compile(source) {
  mod = await WebAssembly.compile(source);
}

function align(v, d) {
  return Math.ceil(v / d) * d;
}

function makeImport(paramTypes, resultTypes, nodeOnly, fn) {
  return function (sp) {
    const args = [];
    let offset = sp + 8;

    const readUint64 = function () {
      offset = align(offset, 8);
      const low = mem.getUint32(offset + 0, true);
      const high = mem.getUint32(offset + 4, true);
      if (high !== 0) {
        throw new Error("value out of supported range");
      }
      offset += 8;
      return low;
    }

    paramTypes.forEach((t) => {
      switch (t) {
        case "int32":
        case "uint32":
          offset = align(offset, 4);
          const val = mem.getUint32(offset + 0, true);
          offset += 4;
          args.push(val);
          break;
        case "int":
        case "int64":
        case "uintptr":
          args.push(readUint64());
          break;
        case "string":
          const strAddr = readUint64();
          const strLen = readUint64();
          args.push(decoder.decode(new DataView(mem.buffer, strAddr, strLen)));
          break;
        case "[]byte":
          const sliceArray = readUint64();
          const sliceLen = readUint64();
          readUint64(); // cap
          args.push(new Uint8Array(mem.buffer, sliceArray, sliceLen));
          break;
        default:
          throw new Error("unknown type: " + t);
      }
    });

    let results = [];
    let errno = 0;
    if (nodeOnly && !isNode) {
      errno = ENOSYS;
    } else {
      try {
        results = fn.apply(null, args);
      } catch (e) {
        if (!e.errno) {
          throw e;
        }
        errno = -e.errno;
      }
    }

    resultTypes.forEach((t, i) => {
      switch (t) {
        case "int32":
          offset = align(offset, 4);
          mem.setUint32(offset, results[i] || 0, true);
          offset += 4;
          break;
        case "int":
        case "int64":
        case "uintptr":
          offset = align(offset, 8);
          const r = results[i] || 0;
          mem.setUint32(offset + 0, r, true);
          mem.setUint32(offset + 4, r / 4294967295, true);
          offset += 8;
          break;
        case "Errno":
          offset = align(offset, 8);
          mem.setUint32(offset + 0, results[i] || errno, true);
          mem.setUint32(offset + 4, 0, true);
          offset += 8;
          break;
        default:
          throw new Error("unknown type: " + t);
      }
    });
  };
}

function setFileStats(st, stats) {
  mem.setUint32(st + 0, stats.dev, true); // Dev int64
  mem.setUint32(st + 8, stats.ino, true); // Ino uint64
  mem.setUint32(st + 16, stats.mode, true); // Mode uint32
  mem.setUint32(st + 20, stats.nlink, true); // Nlink uint32
  mem.setUint32(st + 24, stats.uid, true); // Uid uint32
  mem.setUint32(st + 28, stats.gid, true); // Gid uint32
  mem.setUint32(st + 32, stats.rdev, true); // Rdev int64
  mem.setUint32(st + 40, stats.size, true); // Size int64
  mem.setUint32(st + 48, stats.blksize, true); // Blksize int32
  mem.setUint32(st + 52, stats.blocks, true); // Blocks int32
  // TODO
  mem.setUint32(st + 56, 0, true); // Atime int64
  mem.setUint32(st + 64, 0, true); // AtimeNsec int64
  mem.setUint32(st + 72, 0, true); // Mtime int64
  mem.setUint32(st + 80, 0, true); // MtimeNsec int64
  mem.setUint32(st + 88, 0, true); // Ctime int64
  mem.setUint32(st + 96, 0, true); // CtimeNsec int64
}

async function run() {
  if (!isNode) {
    outputBuf = "";
    console.clear();
  }

  let importObject = {
    js: {
      // func wasmexit(code int32)
      "runtime.wasmexit": makeImport(["int32"], [], false, function (code) {
        if (isNode) {
          process.exit(code);
        } else {
          if (code !== 0) {
            console.warn("exit code:", code);
          }
        }
      }),

      // func write(fd uintptr, p unsafe.Pointer, n int32) int32
      "runtime.write": makeImport(["uintptr", "uintptr", "int32"], ["int32"], false, function (fd, p, n) {
        if (isNode) {
          nodeFS.writeSync(fd, Buffer.from(mem.buffer), p, n);
        } else {
          writeToConsole(new Uint8Array(mem.buffer, p, n));
        }
        return [n];
      }),

      // func open(path string, openmode int, perm uint32) (fd int, errno Errno)
      "syscall.open": makeImport(["string", "int", "uint32"], ["int", "Errno"], true, function (path, openmode, perm) {
        let flags = 0;
        if ((openmode & 1) !== 0) { flags |= nodeFS.constants.O_WRONLY }
        if ((openmode & 2) !== 0) { flags |= nodeFS.constants.O_RDWR }
        if ((openmode & 0100) !== 0) { flags |= nodeFS.constants.O_CREAT }
        if ((openmode & 01000) !== 0) { flags |= nodeFS.constants.O_TRUNC }
        if ((openmode & 02000) !== 0) { flags |= nodeFS.constants.O_APPEND }
        if ((openmode & 0200) !== 0) { flags |= nodeFS.constants.O_EXCL }
        if ((openmode & 04000) !== 0) { flags |= nodeFS.constants.O_NONBLOCK }
        if ((openmode & 010000) !== 0) { flags |= nodeFS.constants.O_SYNC }

        const fd = nodeFS.openSync(path, flags, perm);
        if (nodeFS.fstatSync(fd).isDirectory()) {
          dirContents[fd] = nodeFS.readdirSync(path);
        }
        return [fd, 0];
      }),

      // func close(fd int)
      "syscall.close": makeImport(["int"], ["Errno"], true, function (fd) {
        delete (dirContents[fd]);
        nodeFS.closeSync(fd);
        return [0];
      }),

      // func read(fd int, b []byte) (int)
      "syscall.read": makeImport(["int", "[]byte"], ["int", "Errno"], true, function (fd, b) {
        const n = nodeFS.readSync(fd, b, 0, b.length);
        return [n, 0];
      }),

      // func write(fd int, b []byte) (int)
      "syscall.write": makeImport(["int", "[]byte"], ["int", "Errno"], false, function (fd, b) {
        if (isNode) {
          let n = 0;
          while(n < b.length) {
            try {
              n += nodeFS.writeSync(fd, b, n, b.length - n);
            } catch (e) {
              if (e.code !== "EAGAIN") {
                return [n, -e.errno];
              }
            }
          }
          return [n, 0];
        } else {
          if (fd === 1 || fd === 2) {
            writeToConsole(b);
          }
          return [b.length, 0];
        }
      }),

      // func pread(fd int, b []byte, offset int64) (int, Errno)
      "syscall.pread": makeImport(["int", "[]byte", "int64"], ["int", "Errno"], true, function (fd, b, offset) {
        const n = nodeFS.readSync(fd, b, 0, b.length, offset);
        return [n, 0];
      }),

      // func pwrite(fd int, b []byte, offset int64) (int, Errno)
      "syscall.pwrite": makeImport(["int", "[]byte", "int64"], ["int", "Errno"], true, function (fd, b, offset) {
        const n = nodeFS.writeSync(fd, b, 0, b.length, offset);
        return [n, 0];
      }),

      // func stat(path string, st *Stat_t)
      "syscall.stat": makeImport(["string", "uintptr"], ["Errno"], true, function (path, st) {
        setFileStats(st, nodeFS.statSync(path));
        return [0];
      }),

      // func lstat(path string, st *Stat_t)
      "syscall.lstat": makeImport(["string", "uintptr"], ["Errno"], true, function (path, st) {
        setFileStats(st, nodeFS.lstatSync(path));
        return [0];
      }),

      // func fstat(fd int, st *Stat_t) error
      "syscall.fstat": makeImport(["int", "uintptr"], ["Errno"], true, function (fd, st) {
        setFileStats(st, nodeFS.fstatSync(fd));
        return [0];
      }),

      // func mkdir(path string, perm uint32)
      "syscall.mkdir": makeImport(["string", "uint32"], ["Errno"], true, function (path, perm) {
        nodeFS.mkdirSync(path, perm);
        return [0];
      }),

      // func symlink(path, link string)
      "syscall.symlink": makeImport(["string", "string"], ["Errno"], true, function (path, link) {
        nodeFS.symlinkSync(path, link);
        return [0];
      }),

      // func readDirent(fd int, buf []byte) (int, Errno)
      "syscall.readDirent": makeImport(["int", "[]byte"], ["int", "Errno"], true, function (fd, buf) {
        let n = 0;
        const names = dirContents[fd];
        const dv = new DataView(buf.buffer, buf.byteOffset, buf.byteLength);
        while (names.length !== 0) {
          const name = names.shift();
          if (n + 2 + name.length > buf.byteLength) {
            names.unshift(name);
            break;
          }
          dv.setUint16(n, 2 + name.length, true);
          buf.set(encoder.encode(name), n + 2);
          n += 2 + name.length;
        }
        return [n, 0];
      }),

      // func time_now() (sec int64, nsec int32, mono int64)
      "time.now": makeImport([], ["int64", "int32", "int64"], false, function () {
        const msec = (new Date).getTime();
        const nt = nanotime();
        return [msec / 1000, (msec % 1000) * 1000000, nt];
      }),

      // func nanotime() int64
      "runtime.nanotime": makeImport([], ["int64"], false, function () {
        return [nanotime()];
      }),

      // func readRand(b []byte)
      "crypto/rand.readRand": makeImport(["[]byte"], [], false, function (b) {
        if (isNode) {
          nodeCrypto.randomFillSync(b);
        } else {
          crypto.getRandomValues(b);
        }
        return [];
      }),

      "trace": function (pcF, pcB, sp) {
        console.log(
          `trace`,
          String(sp).padStart(10),
          String(pcF).padStart(10),
          String(pcB).padStart(10),
          sp <= prevSP ? "->" : "<-",
          names[pcF],
        );
        prevSP = sp;
      },
    }
  };

  const nameSec = new SimpleStream(new Uint8Array(WebAssembly.Module.customSections(mod, "name")[0]));
  while (!nameSec.atEnd()) {
    const nameType = nameSec.readByte();
    const namePayloadLen = nameSec.readUleb128();
    const namePayloadData = new SimpleStream(nameSec.read(namePayloadLen));
    if (nameType === 1) { // function names
      const count = namePayloadData.readUleb128();
      for (let i = 0; i < count; i++) {
        const index = namePayloadData.readUleb128();
        const nameLen = namePayloadData.readUleb128();
        const nameStr = String.fromCharCode.apply(null, namePayloadData.read(nameLen));
        names[index] = nameStr;
      }
    }
  }

  const inst = await WebAssembly.instantiate(mod, importObject);
  mem = new DataView(inst.exports.mem.buffer);

  let offset = 1024;

  const strPtr = (str) => {
    let ptr = offset;
    new Uint8Array(mem.buffer, offset, str.length + 1).set(encoder.encode(str + "\0"));
    offset += str.length + (8 - (str.length % 8));
    return ptr;
  };

  const argc = args.length;

  const argvPtrs = [];
  args.forEach((arg) => {
    argvPtrs.push(strPtr(arg));
  });

  const env = isNode ? process.env : {};
  const keys = Object.keys(env).sort();
  argvPtrs.push(keys.length);
  keys.forEach((key) => {
    argvPtrs.push(strPtr(`${key}=${env[key]}`));
  });

  const argv = offset;
  argvPtrs.forEach((ptr) => {
    mem.setUint32(offset, ptr, true);
    mem.setUint32(offset + 4, 0, true);
    offset += 8;
  });

  try {
    inst.exports.run(argc, argv, trace);
  } catch (err) {
    console.error(err);
  }
  if (isNode) {
    process.exit(1);
  }
}

class SimpleStream {
  constructor(array) {
    this.array = array;
    this.offset = 0;
  }

  atEnd() {
    return this.offset >= this.array.length;
  }

  read(len) {
    const a = this.array.subarray(this.offset, this.offset + len);
    this.offset += len;
    return a;
  }

  readByte() {
    const b = this.array[this.offset];
    this.offset++;
    return b;
  }

  readUleb128() {
    let value = 0;
    let shift = 0;
    while (true) {
      let byte = this.readByte();
      value |= (byte & 0x7F) << shift;
      if ((byte & 0x80) === 0) { break; }
      shift += 7;
    }
    return value;
  }
}

function writeToConsole(b) {
  outputBuf += decoder.decode(b);

  const nl = outputBuf.lastIndexOf("\n");
  if (nl != -1) {
    console.log(outputBuf.substr(0, nl));
    outputBuf = outputBuf.substr(nl + 1);
  }
}
