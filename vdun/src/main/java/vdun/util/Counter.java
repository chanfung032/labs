package vdun.util;

import java.util.ArrayList;
import java.util.Comparator;
import java.util.HashMap;
import java.util.List;

public class Counter<T> extends HashMap<T, Integer> {

    public void increase(T obj) {
        increase(obj, 1);
    }

    public void increase(T obj, int n) {
        put(obj, getOrDefault(obj, 0) + n);
    }

    public List<Entry<T, Integer>> getMostCommon() {
        return getMostCommon(size());
    }

    public List<Entry<T, Integer>> getMostCommon(int n) {
        List<Entry<T, Integer>> c = new ArrayList<Entry<T, Integer>>(entrySet());
        c.sort(new Comparator<Entry<T, Integer>>() {
            public int compare(Entry<T, Integer> l, Entry<T, Integer> r) {
                return (r.getValue()).compareTo(l.getValue());
            }
        });
        return size() > n ? c.subList(0, Math.min(c.size(), n)) : c;
    }
}