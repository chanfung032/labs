package vdun.util;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Map.Entry;
import static vdun.util.Number.*;

public class Feature {
    private SlidingWindowCounterMonad pv;
    private Object[] counters;
    private static String[] features = new String[]{"path", "user_agent", "referer", "status", "request_pattern"};

    public Feature(int featureWindowLength) {
        pv = new SlidingWindowCounterMonad(featureWindowLength);

        counters = new Object[features.length];
        for (int i = 0; i < features.length; i++) {
            counters[i] = new SlidingWindowCounter<String>(featureWindowLength);
        }
    }

    public void update(Request request) {
        pv.increase(1);

        int i = 0;
        for (String feature : features) {
            ((SlidingWindowCounter<String>)counters[i++]).increase(request.get(feature), 1);
        }
    }

    public boolean empty() {
        return pv.getCount() == 0;
    }

    public Map<String, Object> getSummaryThenAdvance() {
        assert(!empty());

        Map<String, Object> summary = new HashMap<String, Object>();

        int pv = this.pv.getCountThenAdvance();
        summary.put("pv", pv);

        for (int i = 0; i < features.length; i++) {
            String feature = features[i];
            SlidingWindowCounter<String> counter = (SlidingWindowCounter<String>)counters[i];
            List<Entry<String, Integer>> data = counter.getMostCommon(5);

            class MostCommon {
                public String name;
                public int pv;
                public float ratio;

                public String toString() {
                    return String.format("name: %s, pv: %d, ration: %.3f", name, pv, ratio);
                }
            }

            MostCommon[] mostCommon = new MostCommon[data.size()];
            List<Double> ratios = new ArrayList<Double>(data.size());
            for (int j = 0; j < mostCommon.length; j++) {
                Entry<String, Integer> e = data.get(j);
                mostCommon[j] = new MostCommon();
                mostCommon[j].name = e.getKey();
                mostCommon[j].pv = e.getValue();
                mostCommon[j].ratio = (float)e.getValue() / pv;
                ratios.add(j, (double)mostCommon[j].ratio);
            }
            summary.put("most_" + feature, Arrays.asList(mostCommon));
            summary.put("pv_most_" + feature, mostCommon[0].pv);
            summary.put("ratio_most_" + feature, mostCommon[0].ratio);
            summary.put("ratio_std_" + feature, std(ratios));
            summary.put("uniq_" + feature, counter.size());

            counter.advanceWindow();
        }

        return summary;
    }
}