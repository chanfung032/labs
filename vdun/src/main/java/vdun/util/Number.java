package vdun.util;

import java.util.List;

public class Number {
    public static double std(List<Double> v) {
        double avgVal = avg(v);
        double retval = 0;
        for (double vv: v) {
            retval += (vv-avgVal) * (vv-avgVal);
        }
        return (double)Math.sqrt(retval / v.size());
    }

    public static double avg(List<Double> v) {
        double sum = 0;
        for (double vv: v) {
            sum += vv;
        }
        return sum / v.size();
    }
}