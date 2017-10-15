package jared.livepark;

import android.support.v7.app.AppCompatActivity;
import android.os.Bundle;
import android.util.Log;
import android.view.View;

import jared.livepark.Models.HttpGetRequest;

public class MyCarActivity extends AppCompatActivity {

    private static final String BEACON_URL = HttpGetRequest.SERVER_ADDRESS + "Beacon";

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_my_car);
    }

    public void beacon(View view) {
        try {
            new HttpGetRequest().execute(BEACON_URL);
            Log.d("TAG", "Success");
        } catch (Exception e) {
            Log.d("TAG", e.toString());
        }
    }
}
