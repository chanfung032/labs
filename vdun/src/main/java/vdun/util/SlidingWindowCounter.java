package vdun.util;

import java.io.Serializable;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Comparator;
import java.util.Set;
import java.util.Map.Entry;

public class SlidingWindowCounter<T> implements Serializable {
    private int headSlot;
    private int windowLength;
    private final Map<T, int[]> objToCounts = new HashMap<T, int[]>();

    public SlidingWindowCounter(int windowLength) {
        if (windowLength < 2) {
            throw new IllegalArgumentException("Window length must be at least two, you requested (" 
                + windowLength + ")");
        }
        this.headSlot = 0;
        this.windowLength = windowLength;
    }

    public void increase(T obj, int n) {
        int[] counts = objToCounts.get(obj);
        if (counts == null) {
            counts = new int[windowLength + 1];
            objToCounts.put(obj, counts);
        }
        counts[headSlot] += n;
    }

    public int size() {
        return objToCounts.size();
    }

    public Map<T, Integer> getCounts() {
        Map<T, Integer> result = new HashMap<T, Integer>();
        for (Map.Entry<T, int[]> e : objToCounts.entrySet()) {
            result.put(e.getKey(), e.getValue()[windowLength] + e.getValue()[headSlot]);
        }
        return result;
    }

    public void advanceWindow() {
        int tailSlot = (headSlot + 1) % windowLength;
        
        Set<T> objToRemove = new HashSet<T>();
        for (Map.Entry<T, int[]> e: objToCounts.entrySet()) {
            int[] counts = e.getValue();
            counts[windowLength] -= counts[tailSlot] - counts[headSlot];
            counts[tailSlot] = 0;
            if (counts[windowLength] == 0) {
                objToRemove.add(e.getKey());
            }
        }

        for (T obj: objToRemove) {
            objToCounts.remove(obj);
        }

        headSlot = tailSlot;
    }

    public Map<T, Integer> getCountsThenAdvanceWindow() {
        Map<T, Integer> counts = this.getCounts();
        this.advanceWindow();
        return counts;
    }

    public List<Entry<T, Integer>> getMostCommon() {
        return getMostCommon(objToCounts.size());
    }

    public List<Entry<T, Integer>> getMostCommon(int n) {
        List<Entry<T, Integer>> c = new ArrayList<Entry<T, Integer>>(getCounts().entrySet());
        c.sort(new Comparator<Entry<T, Integer>>() {
            public int compare(Entry<T, Integer> l, Entry<T, Integer> r) {
                return (r.getValue()).compareTo(l.getValue());
            }
        });
        return objToCounts.size() > n ? c.subList(0, Math.min(c.size(), n)) : c;
    }
}