package jared.livepark;

import android.Manifest;
import android.content.pm.PackageManager;
import android.graphics.Bitmap;
import android.graphics.Canvas;
import android.graphics.drawable.Drawable;
import android.location.Location;
import android.location.LocationManager;
import android.content.Context;
import android.location.Criteria;
import android.support.v4.app.FragmentActivity;
import android.os.Bundle;
import android.support.v4.content.ContextCompat;
import android.util.Log;

import com.google.android.gms.maps.CameraUpdateFactory;
import com.google.android.gms.maps.GoogleMap;
import com.google.android.gms.maps.OnMapReadyCallback;
import com.google.android.gms.maps.SupportMapFragment;
import com.google.android.gms.maps.model.BitmapDescriptor;
import com.google.android.gms.maps.model.BitmapDescriptorFactory;
import com.google.android.gms.maps.model.LatLng;
import com.google.android.gms.maps.model.Marker;
import com.google.android.gms.maps.model.MarkerOptions;
import com.google.android.gms.maps.model.Polygon;
import com.google.android.gms.maps.model.PolygonOptions;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.io.IOException;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;

import cn.pedant.SweetAlert.SweetAlertDialog;
import jared.livepark.Models.ParkingLot;
import jared.livepark.Models.HttpGetRequest;
import jared.livepark.Models.ParkingSpot;

public class MapsActivity extends FragmentActivity implements OnMapReadyCallback,
        GoogleMap.OnMarkerClickListener, SweetAlertDialog.OnSweetClickListener {

    private static final String SERVER_ADDRESS = "http://172.20.10.6:8080/";
    private static final String LOT_GET_URL = SERVER_ADDRESS + "LotInfo";
    private static final String REQ_SPOT_URL = SERVER_ADDRESS + "ReqSpot";

    private GoogleMap mMap;

    private HashMap<String, ParkingLot> lotMap;

    private String previousMarkerClick = "";

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
        if (ContextCompat.checkSelfPermission(this,
                Manifest.permission.ACCESS_FINE_LOCATION)
                == PackageManager.PERMISSION_GRANTED) {
            mMap.setMyLocationEnabled(true);
        } else {
            throw new java.lang.RuntimeException("Location permissions disabled");
        }
        Location location = getCurrentLocation();
        mMap.moveCamera(CameraUpdateFactory.newLatLngZoom(
                new LatLng(location.getLatitude(), location.getLongitude()),
                14
        ));
        List<ParkingLot> lots = queryLots();
        lotMap = new HashMap<>();
        for (ParkingLot lot : lots) {
            mMap.addMarker(new MarkerOptions()
                    .position(lot.getEntrance())
                    .title(lot.getTitle()));
            lotMap.put(lot.getTitle(), lot);
        }
        LatLng[] fakeLots = {
                new LatLng(33.763741, -84.393911),
                new LatLng(33.786590, -84.401224),
                new LatLng(33.776174, -84.388864),
                new LatLng(33.785734, -84.386032),
                new LatLng(33.773088, -84.383438)
        };
        for (LatLng lotPos : fakeLots) {
            mMap.addMarker(new MarkerOptions()
                    .position(lotPos)
                    .title("Unsupported Lot"));
        }

        mMap.setOnMarkerClickListener(this);
    }

    public Location getCurrentLocation() {
        if (ContextCompat.checkSelfPermission(this,
                Manifest.permission.ACCESS_FINE_LOCATION)
                == PackageManager.PERMISSION_GRANTED) {
            LocationManager locationManager = (LocationManager)
                    getSystemService(Context.LOCATION_SERVICE);
            Criteria criteria = new Criteria();
            return locationManager.getLastKnownLocation(locationManager
                    .getBestProvider(criteria, false));
        } else {
            return null;
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

    @Override
    public boolean onMarkerClick(final Marker marker) {
        if (marker.getTitle().equals("Your Space") || marker.getTitle().equals("Unsupported Lot")) {
            return false;
        } else if (marker.getTitle().equals("Entrance Location")) {
            return true;
        }
        Log.d("TAG", "Marker clicked");
        mMap.animateCamera(CameraUpdateFactory.newLatLngZoom(marker.getPosition(), 18));
        marker.showInfoWindow();
        ParkingLot lot = lotMap.get(marker.getTitle());
        PolygonOptions rectOptions = new PolygonOptions();
        for (LatLng fencePoint : lot.getFence()) {
            rectOptions.add(fencePoint);
        }
        rectOptions.fillColor(0x5fff0000);
        rectOptions.strokeWidth(0);
        if (lot.getTitle().equals(previousMarkerClick)) {
            SweetAlertDialog pDialog = new SweetAlertDialog(this, SweetAlertDialog.NORMAL_TYPE);
            pDialog.setTitleText(lot.getTitle());
            String dialogContent = String.format(
                    "Available spots: %d\nPrice: $%1.2f/hour\nWould you like to park here?",
                    lot.getAvailableSpots(), lot.getPrice());
            pDialog.setContentText(dialogContent);
            pDialog.setConfirmText("Park");
            pDialog.setCancelText("Cancel");
            pDialog.setConfirmClickListener(this);
            pDialog.showCancelButton(true);
            pDialog.show();
        }
        mMap.addPolygon(rectOptions);
        previousMarkerClick = lot.getTitle();
        return true;
    }

    @Override
    public void onClick(SweetAlertDialog sweetAlertDialog) {
        mMap.clear();
        ParkingLot lot = lotMap.get(sweetAlertDialog.getTitleText());
        sweetAlertDialog.cancel();
        String url = String.format(REQ_SPOT_URL + "?lat=%1.4f&long=%1.4f",
                lot.getEntrance().latitude, lot.getEntrance().longitude);
        Log.d("URL", url);
        ParkingSpot spot;
        try {
            String json = new HttpGetRequest().execute(url).get();
            Log.d("JSON", json);
            spot = parseSpotJson(json);
        } catch (Exception e) {
            Log.d("TAG", e.toString());
            return;
        }
        mMap.addMarker(new MarkerOptions()
                .position(spot.getCoordinates())
                .title("Your Space"));
        mMap.addMarker(new MarkerOptions()
                .position(lot.getEntrance())
                .title("Entrance Location")
                .icon(bitmapDescriptorFromVector(getApplicationContext(), R.drawable.entrance_icon)));
        PolygonOptions rectOptions = new PolygonOptions();
        for (LatLng fencePoint : lot.getFence()) {
            rectOptions.add(fencePoint);
        }
        rectOptions.fillColor(0x5fff0000);
        rectOptions.strokeWidth(0);
        mMap.addPolygon(rectOptions);

        SweetAlertDialog info = new SweetAlertDialog(this, SweetAlertDialog.SUCCESS_TYPE);
        info.setTitleText("Success");
        info.setContentText("Spot " + spot.getName() + " has been reserved.");
        info.show();
    }

    private BitmapDescriptor bitmapDescriptorFromVector(Context context, int vectorResId) {
        Drawable vectorDrawable = ContextCompat.getDrawable(context, vectorResId);
        vectorDrawable.setBounds(0, 0, vectorDrawable.getIntrinsicWidth(), vectorDrawable.getIntrinsicHeight());
        Bitmap bitmap = Bitmap.createBitmap(vectorDrawable.getIntrinsicWidth(), vectorDrawable.getIntrinsicHeight(), Bitmap.Config.ARGB_8888);
        Canvas canvas = new Canvas(bitmap);
        vectorDrawable.draw(canvas);
        return BitmapDescriptorFactory.fromBitmap(bitmap);
    }

    public ParkingSpot parseSpotJson(String json) throws JSONException {
        JSONObject spotJson = new JSONObject(json);
        return new ParkingSpot(
                spotJson.getString("Name"),
                spotJson.getDouble("Latitude"),
                spotJson.getDouble("Longitude")
        );
    }
}
