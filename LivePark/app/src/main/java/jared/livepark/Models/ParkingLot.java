package jared.livepark.Models;

import com.google.android.gms.maps.model.LatLng;

import java.util.List;

public class ParkingLot {
    private List<LatLng> fence;
    private LatLng entrance;
    private String title;
    private double price;
    private int availableSpots;
    private int totalSpots;

    public ParkingLot(List<LatLng> fence, LatLng entrance, String title,
                      double price, int availableSpots, int totalSpots) {
        this.fence = fence;
        this.entrance = entrance;
        this.title = title;
        this.price = price;
        this.availableSpots = availableSpots;
        this.totalSpots = totalSpots;
    }

    public LatLng getEntrance() {
        return entrance;
    }

    public List<LatLng> getFence() {
        return fence;
    }

    public String getTitle() {
        return title;
    }

    public double getPrice() {
        return price;
    }

    public int getAvailableSpots() {
        return availableSpots;
    }

    public int getTotalSpots() {
        return totalSpots;
    }
}
