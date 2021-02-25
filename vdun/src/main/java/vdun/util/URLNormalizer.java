package vdun.util;

import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.TreeMap;
import java.util.Map.Entry;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class URLNormalizer {
    private static final Logger LOG = LoggerFactory.getLogger(URLNormalizer.class);
    private static final Pattern QS_PATTERN = Pattern.compile("([^=]+)=([^&]*)&?");
    private static final Pattern EXT_PATTERN = Pattern.compile("\\.\\w+$");
    private Set<String> K = new HashSet<String>();
    private Set<String> V = new HashSet<String>();

    public void fit(List<String> urls) {
        List<Map<String, String>> U = new ArrayList<Map<String,String>>();
        for (String url: urls) {
            U.add(toKeyValuePair(url));
        }

        Counter<String> C_keys = new Counter<String>();
        for (Map<String, String> u : U) {
            for (String k : u.keySet()) {
                C_keys.increase(k);
            }
        }
        for (Entry<String, Integer> e : C_keys.entrySet()) {
            if (e.getKey().startsWith("\0") || e.getValue() > 1) {
                K.add(e.getKey());
            }
        }
        LOG.info("K = {}", K);

        Counter<String> C_values = new Counter<String>();
        for (Map<String, String> u : U) {
            for (String v : u.values()) {
                C_values.increase(v);
            }
        }
        List<Entry<String, Integer>> most = C_values.getMostCommon();
        int pos = 0;
        double maxDec = 0;
        for (int i = 0; i < most.size() - 1; i++) {
            int l = most.get(i).getValue();
            int r = most.get(i+1).getValue();
            if (r == 1) {
                break;
            }
            double dec = Math.log((double)l) - Math.log((double)r);
            if (dec >= maxDec) {
                pos = i; maxDec = dec;
            }
        }
        for (int i = 0; i < pos+1; i++) {
            V.add(most.get(i).getKey());
        }
        LOG.info("V = {}", V);

        V.addAll(extractAllKeywords(U, K));
        LOG.info("V = {}", V);
    }

    public String transform(String url) {
        assert K != null & V != null;
        Map<String, String> kvs = toKeyValuePair(url);
        LOG.info("kvs = {}", kvs);

        List<String> paths = new ArrayList<String>();
        int i;
        for (i = 0; ; i++) {
            String v = kvs.get(String.format("\0%d", i));
            if (v == null) {
                break;
            }
            paths.add("/" + (V.contains(v) ? v : "*"));
        }
        if (paths.size() > 0 && paths.get(paths.size()-1).equals("/*")) {
            // 始终保留路径的扩展名
            String last = paths.get(paths.size()-1);
            Matcher m = EXT_PATTERN.matcher(kvs.get(String.format("\0%d", i-1)));
            if (m.find()) {
                paths.set(paths.size()-1, last + m.group());
            }
        }
        String normalizedURL = String.join("", paths);

        List<String> qs = new ArrayList<String>();
        for (Entry<String, String> e : kvs.entrySet()) {
            if (!e.getKey().startsWith("\0") && (K.size() == 0 || K.contains(e.getKey()))) {
                if (V.contains(e.getValue())) {
                    qs.add(String.format("%s=%s", e.getKey(), e.getValue()));
                } else {
                    qs.add(String.format("%s=*", e.getKey()));
                }
            }
        }
        if (qs.size() > 0) {
            normalizedURL += "?" + String.join("&", qs);
        }

        return normalizedURL;
    }

    public Object[] getParams() {
        return new Object[]{K, V};
    }

    public void setParams(Set<String>[] params) {
        assert params.length == 2;
        K = params[0]; V = params[1];
        LOG.info("set params, K: {}, V: {}", K, V);
    }

    private Map<String, String> toKeyValuePair(String url) {
        Map<String, String> pairs = new TreeMap<String,String>();
        String[] p = url.split("\\?", 2);
        String[] paths = p[0].split("/");
        for (int i = 0, j = 0; i < paths.length; i++) {
            if (paths[i].length() != 0) {
                pairs.put(String.format("\0%d", j++), paths[i]);
            }
        }
        if (p.length > 1) {
            Matcher m = QS_PATTERN.matcher(p[1]);
            while (m.find()) {
                pairs.put(m.group(1), m.group(2));
            }
        }
        LOG.info("pairs = {}", pairs);
        return pairs;
    }

    private Set<String> extractAllKeywords(List<Map<String, String>> U, Set<String> K) {
        if (K.size() == 0) {
            return extractKeywords(U);
        }

        int N = U.size();

        // \0 表示值不存在，\1 表示这是一个通配 *
        String K_star = null;
        double minHk = Double.POSITIVE_INFINITY;
        for (String k : K) {
            Counter<String> C = new Counter<String>();
            for (Map<String, String> u : U) {
                C.increase(u.getOrDefault(k, "\1"));
            }
            double Hk = 0;
            for (Integer e : C.values()) {
                double ratio = ((double)e) / N;
                Hk += -ratio * Math.log(ratio);
            }
            if (Hk < minHk) {
                minHk = Hk; K_star = k;
                LOG.info("k: {} hk: {}", k, Hk);
            }
        }
        LOG.info("k_start {}", K_star);

        Set<String> V_s = new HashSet<String>();
        for (Map<String, String> u : U) {
            String v = u.getOrDefault(K_star, "\0");
            if (V.contains(v) || v.equals("\0")) {
                V_s.add(v);
            } else {
                V_s.add("\1");
            }
        }

        if (V_s.size() == 1 && V_s.contains("\1")) {
            return extractKeywords(U);
        }

        Set<String> keywords = new HashSet<String>();
        Set<String> K_i = new HashSet<String>(K);
        K_i.remove(K_star);
        for (String v : V_s) {
            List<Map<String, String>> U_i = new ArrayList<Map<String, String>>();
            if (v.equals("\1")) {
                for (Map<String, String> u : U) {
                    if (!V_s.contains(u.getOrDefault(K_star, "\0"))) {
                        U_i.add(u);
                    }
                }
            } else {
                for (Map<String, String> u : U) {
                    if (u.getOrDefault(K_star, "\0").equals(v)) {
                        U_i.add(u);
                    }
                }
            }
            keywords.addAll(extractAllKeywords(U_i, K_i));
        }

        return keywords;
    }

    private Set<String> extractKeywords(List<Map<String, String>> U) {
        LOG.info("extract {}", U);
        Set<String> K = new HashSet<String>();
        for (Map<String, String> u : U) {
            K.addAll(u.keySet());
        }

        Set<String> keywords = new HashSet<String>();
        for (String k : K) {
            Set<String> V = new HashSet<String>();
            for (Map<String, String> u : U) {
                String v = u.get(k);
                if (v != null) {
                    V.add(v);
                }
            }

            if (V.size() < Math.max(2, Math.min(U.size()/5, 10))) {
                keywords.addAll(V);
            }
        }

        return keywords;
    }
}