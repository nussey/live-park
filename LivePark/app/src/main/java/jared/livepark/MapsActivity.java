package jared.livepark;

import android.Manifest;
import android.content.pm.PackageManager;
import android.location.Location;
import android.location.LocationManager;
import android.content.Context;
import android.location.Criteria;
import android.support.v4.app.FragmentActivity;
import android.os.Bundle;
import android.support.v4.content.ContextCompat;
import android.util.Log;
import android.widget.Toast;

import com.google.android.gms.maps.CameraUpdateFactory;
import com.google.android.gms.maps.GoogleMap;
import com.google.android.gms.maps.OnMapReadyCallback;
import com.google.android.gms.maps.SupportMapFragment;
import com.google.android.gms.maps.model.LatLng;
import com.google.android.gms.maps.model.MarkerOptions;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.URL;
import java.util.ArrayList;
import java.util.List;

import jared.livepark.Models.ParkingLot;
import jared.livepark.Models.HttpGetRequest;

public class MapsActivity extends FragmentActivity implements OnMapReadyCallback {

    private static final String LOT_GET_URL = "http://172.20.10.6:8080/LotInfo";

    private GoogleMap mMap;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_maps);
        // Obtain the SupportMapFragment and get notified when the map is ready to be used.
        SupportMapFragment mapFragment = (SupportMapFragment) getSupportFragmentManager()
                .findFragmentById(R.id.map);
        mapFragment.getMapAsync(this);
    }

    @Override
    public void onMapReady(GoogleMap googleMap) {
        mMap = googleMap;
        if (ContextCompat.checkSelfPermission(this, Manifest.permission.ACCESS_FINE_LOCATION) == PackageManager.PERMISSION_GRANTED) {
            mMap.setMyLocationEnabled(true);
        } else {
            throw new java.lang.RuntimeException("Location permissions disabled");
        }
        LocationManager locationManager = (LocationManager)
                getSystemService(Context.LOCATION_SERVICE);
        Criteria criteria = new Criteria();

        Location location = locationManager.getLastKnownLocation(locationManager
                .getBestProvider(criteria, false));
        mMap.moveCamera(CameraUpdateFactory.newLatLngZoom(
                new LatLng(location.getLatitude(), location.getLongitude()),
                13
        ));
        List<ParkingLot> lots = queryLots();
        for (ParkingLot lot : lots) {
            mMap.addMarker(new MarkerOptions().position(lot.getEntrance()).title(lot.getTitle()));
        }
    }

    public List<ParkingLot> queryLots() {
        // make parking lot query
        try {
            return parseLotJson(new HttpGetRequest().execute(LOT_GET_URL).get());
        } catch (Exception e) {
            return null;
        }
    }

    public List<ParkingLot> parseLotJson(String json) {
        List<ParkingLot> lots = new ArrayList<ParkingLot>();
        try {
            JSONObject lotJson = new JSONObject(json);

            JSONArray fenceJson = lotJson.getJSONArray("GeoFence");
            List<LatLng> fence = new ArrayList<LatLng>();
            for (int j = 0; j < fenceJson.length(); j++) {
                JSONObject fencePointJson = fenceJson.getJSONObject(j);
                fence.add(new LatLng(fencePointJson.getDouble("Latitude"),
                        fencePointJson.getDouble("Longitude")));
            }

            JSONObject entranceJson = lotJson.getJSONObject("Entrance");
            LatLng entrance = new LatLng(entranceJson.getDouble("Latitude"),
                    entranceJson.getDouble("Longitude"));

            lots.add(new ParkingLot(
                    fence,
                    entrance,
                    lotJson.getString("Name"),
                    lotJson.getDouble("Price"),
                    lotJson.getInt("AvailableSpots"),
                    lotJson.getInt("TotalSpots")
            ));
        } catch (JSONException e) {
            Log.e("MapsActivity", "Error parsing parking lots JSON.");
        }
        return lots;
    }
}
