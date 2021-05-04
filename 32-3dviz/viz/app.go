package main

import (
	"context"
	"embed"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/StefanSchroeder/Golang-Ellipsoid/ellipsoid"
	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/texture"
	"github.com/g3n/engine/window"
	"github.com/uber/h3-go/v3"
)

type App struct {
	wd     string
	v      *Viz
	geo    ellipsoid.Ellipsoid
	hexMap map[h3.H3Index]*HexNode
	a      *app.Application
	cam    *camera.Camera
	scene  *core.Node
	earth  *graphic.Mesh
	hexes  *core.Node
}

func newApp(v *Viz) *App {
	geo := ellipsoid.Init("WGS84", ellipsoid.Degrees, ellipsoid.Meter, ellipsoid.LongitudeIsSymmetric, ellipsoid.BearingIsSymmetric)
	return &App{
		v:      v,
		geo:    geo,
		hexMap: make(map[h3.H3Index]*HexNode),
	}
}

func (g *App) run(ctx context.Context) error {
	// get working directory
	var err error
	g.wd, err = os.Getwd()
	if err != nil {
		return err
	}

	// build base app and scene
	g.a = app.App()
	g.scene = core.NewNode()
	gui.Manager().Set(g.scene)
	g.cam = camera.New(1)
	g.cam.SetPosition(0, 0, 3)
	g.scene.Add(g.cam)
	camera.NewOrbitControl(g.cam)
	g.a.Subscribe(window.OnWindowSize, g.resize)
	g.resize("", nil)

	// skybox and lights
	if err := g.addSkyboxAndLights(); err != nil {
		return err
	}
	// earth
	if err := g.addEarth(); err != nil {
		return err
	}
	// sun
	if err := g.addSun(); err != nil {
		return err
	}
	// hexes
	g.addHexes()
	// run
	go func() {
		for {
			ch := time.After(time.Duration(0.5 * float64(time.Second)))
			// update opacity on hexes
			m := make(map[h3.H3Index]int)
			for _, msg := range g.v.messages[:] {
				m[msg.Parent]++
			}
			for k, v := range g.hexMap {
				f := float32(0.0)
				if g, ok := m[k]; ok {
					f = float32(g) / 20.0
				}
				v.mat.SetOpacity(min(f, 0.9))
			}
			select {
			case <-ctx.Done():
				return
			case <-ch:
			}
		}
	}()
	g.a.Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {
		g.a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		rotation := 0.02 * float32(deltaTime.Seconds())
		g.hexes.RotateY(rotation)
		g.earth.RotateY(rotation)
		renderer.Render(g.scene, g.cam)
	})
	return nil
}

func (g *App) resize(n string, ev interface{}) {
	width, height := g.a.GetSize()
	g.a.Gls().Viewport(0, 0, int32(width), int32(height))
	g.cam.SetAspect(float32(width) / float32(height))
}

func (g *App) addSkyboxAndLights() error {
	// ambient and directional lighting
	ambLight := light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.5)
	ambLight.SetIntensity(0.3)
	g.scene.Add(ambLight)
	dirLight := light.NewDirectional(&math32.Color{1, 1, 1}, 0.9)
	dirLight.SetPosition(0, 0, 100)
	g.scene.Add(dirLight)
	return nil
	// skybox
	skybox, err := graphic.NewSkybox(graphic.SkyboxData{
		filepath.Join(g.wd, "viz", "images", "dark-s_"), "jpg",
		[6]string{"px", "nx", "py", "ny", "pz", "nz"},
	})
	if err != nil {
		return err
	}
	g.scene.Add(skybox)
	return nil
}

func (g *App) addEarth() error {
	mat, err := newEarthMaterial(g, &math32.Color{1, 1, 1})
	if err != nil {
		return err
	}
	geom := geometry.NewSphere(0.99, 32, 16)
	g.earth = graphic.NewMesh(geom, mat)
	g.scene.Add(g.earth)
	return nil
}

func (g *App) addSun() error {
	tex, err := newTexture(g.wd, "lensflare0_alpha.png")
	if err != nil {
		return err
	}
	mat := material.NewStandard(&math32.Color{1, 1, 1})
	mat.AddTexture(tex)
	mat.SetTransparent(true)
	sun := graphic.NewSprite(10, 10, mat)
	sun.SetPositionZ(20)
	g.scene.Add(sun)
	return nil
}

func (g *App) addHexes() {
	// hexes
	g.hexes = core.NewNode()
	g.scene.Add(g.hexes)
	var count int
	for _, parent := range h3.GetRes0Indexes() {
		for _, h := range h3.ToChildren(parent, g.v.res) {
			hex := g.newHexNode(h, &math32.Color{1.0, 0.0, 0.0})
			g.hexMap[h] = hex
			g.hexes.Add(hex)
			count++
		}
	}
	log.Printf("hexes: %d", count)
}

func (g *App) newHexNode(h h3.H3Index, color *math32.Color) *HexNode {
	// get coordinates
	ctr := g.coordToVec3(h3.ToGeo(h), true)
	boundary := h3.ToGeoBoundary(h)
	radius := ctr.DistanceTo(g.coordToVec3(boundary[0], true))
	// mesh data
	positions := math32.NewArrayF32(0, 16)
	normals := math32.NewArrayF32(0, 16)
	uvs := math32.NewArrayF32(0, 16)
	indices := math32.NewArrayU32(0, 16)
	// append center and normal
	positions.AppendVector3(ctr)
	normals.AppendVector3(ctr)
	// append center uv coordinate
	centerUV := math32.NewVector2(0.5, 0.5)
	uvs.AppendVector2(centerUV)
	// generate tris
	boundary = append(boundary, boundary[0])
	for _, h := range boundary {
		// Appends vertex position, normal and uv coordinates
		p := g.coordToVec3(h, true)
		positions.AppendVector3(p)
		normals.AppendVector3(ctr)
		uvs.Append((p.X/float32(radius)+1)/2, (p.Y/float32(radius)+1)/2)
	}
	// add indices
	for i := 1; i <= len(boundary); i++ {
		indices.Append(uint32(i), uint32(i)+1, 0)
	}
	// geometry
	d := geometry.NewGeometry()
	d.SetIndices(indices)
	d.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))
	d.AddVBO(gls.NewVBO(normals).AddAttrib(gls.VertexNormal))
	d.AddVBO(gls.NewVBO(uvs).AddAttrib(gls.VertexTexcoord))
	// mesh and node
	mat := material.NewStandard(color)
	mat.SetTransparent(true)
	mat.SetOpacity(0.0)
	mesh := graphic.NewMesh(d, mat)
	node := core.NewNode()
	node.Add(mesh)
	hex := &HexNode{
		Node: node,
		mat:  mat,
	}
	return hex
}

func (g *App) coordToVec3(c h3.GeoCoord, normalize bool) *math32.Vector3 {
	x, y, z := g.geo.ToECEF(c.Latitude, c.Longitude, 0.0)
	vec3 := math32.NewVector3(float32(x), float32(y), float32(z))
	if normalize {
		return vec3.Normalize()
	}
	return vec3
}

type EarthMaterial struct {
	material.Standard
}

func newEarthMaterial(g *App, color *math32.Color) (*EarthMaterial, error) {
	// textures
	texDay, err := newTexture(g.wd, "earth_clouds_big.jpg")
	if err != nil {
		return nil, err
	}
	texSpecular, err := newTexture(g.wd, "earth_spec_big.jpg")
	if err != nil {
		return nil, err
	}
	texNight, err := newTexture(g.wd, "earth_night_big.jpg")
	if err != nil {
		return nil, err
	}
	// shaders
	earthVert, err := shaders.ReadFile("earth.vert")
	if err != nil {
		return nil, err
	}
	earthFrag, err := shaders.ReadFile("earth.frag")
	if err != nil {
		return nil, err
	}
	// build programs
	renderer := g.a.Renderer()
	renderer.AddShader("shaderEarthVertex", string(earthVert))
	renderer.AddShader("shaderEarthFrag", string(earthFrag))
	renderer.AddProgram("shaderEarth", "shaderEarthVertex", "shaderEarthFrag")
	// material
	mat := new(EarthMaterial)
	mat.Standard.Init("shaderEarth", color)
	mat.SetShininess(20)
	mat.AddTexture(texDay)
	mat.AddTexture(texSpecular)
	mat.AddTexture(texNight)
	return mat, nil
}

func newTexture(wd string, paths ...string) (*texture.Texture2D, error) {
	tex, err := texture.NewTexture2DFromImage(filepath.Join(append([]string{wd, "viz", "images"}, paths...)...))
	if err != nil {
		return nil, err
	}
	tex.SetFlipY(false)
	return tex, nil
}

type HexNode struct {
	*core.Node
	mat   *material.Standard
	total int
}

func min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

//go:embed *.vert
//go:embed *.frag
var shaders embed.FS
