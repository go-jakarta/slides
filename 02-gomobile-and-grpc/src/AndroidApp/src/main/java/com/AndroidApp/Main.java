package com.AndroidApp;

import android.app.Activity;
import android.os.Bundle;
import android.widget.Button;
import android.widget.TextView;
import android.view.View;

// IMPORT OMIT
import go.helloclient.Helloclient;
import go.helloclient.Helloclient.HelloClient;
// END OMIT

public class Main extends Activity {
    private static final String addr = "192.168.1.5:8833";

    /** Called when the activity is first created. */
    @Override
    public void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.main);

        final Button button = (Button) findViewById(R.id.sayhello);
        final TextView textView = (TextView) findViewById(R.id.result);
        final StringBuilder sb = new StringBuilder("");

// CLICK OMIT
        // add click listener
        button.setOnClickListener(new View.OnClickListener() {
            public void onClick(View v) {
                try {
                    // set some status text
                    sb.append("Trying " + addr + "\n");
                    textView.setText(sb.toString());

                    // create the Go type instance
                    HelloClient hc = Helloclient.New(addr);

                    // call method
                    String result = hc.SayHello("john");

                    // process result
                    sb.append("Received: \"" + result + "\"\n");
                    textView.setText(sb.toString());

                    // close
                    hc.Shutdown();
                } catch (Exception e) {
                    sb.append("error: " + e + "\n");
                    textView.setText(sb.toString());
                }
            }
        });
// END OMIT
    }
}
