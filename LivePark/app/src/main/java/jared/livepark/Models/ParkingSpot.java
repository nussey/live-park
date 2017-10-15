package jared.livepark.Models;

import com.google.android.gms.maps.model.LatLng;

public class ParkingSpot {
    private String name;
    private LatLng coordinates;

    public ParkingSpot(String name, double latitude, double longitude) {
        this.name = name;
        this.coordinates = new LatLng(latitude, longitude);
    }

    public LatLng getCoordinates() {
        return coordinates;
    }

    public String getName() {
        return name;
    }
}
