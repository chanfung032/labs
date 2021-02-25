package vdun.util;

import java.util.Map;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class ConfigParser {
    private static final Logger LOG = LoggerFactory.getLogger(ConfigParser.class);
    private Map<String, Object> confMap;

    public ConfigParser(Map<String, Object> confMap) {
        this.confMap = confMap;
    }

    public Integer getInteger(String key, Integer default_) {
        return (Integer)get(key, default_, Integer.class);
    }

    public Long getLong(String key, Long default_) {
        return (Long)get(key, default_, Long.class);
    }

    public Boolean getBoolean(String key, boolean default_) {
        return (Boolean)get(key, default_, Boolean.class);
    }

    public String getString(String key, String default_) {
        return (String)get(key, default_, String.class);
    }

    public Object get(String key, Object default_, Class<?> type) {
        String[] keys = key.split("\\.");
        Map<String, Object> m = confMap;
        Object v = null;
        for (int i = 0; ; i++) {
            v = m.get(keys[i]);
            if (v == null) {
                return default_;
            }
            if (i == keys.length - 1) {
                break;
            } else if (v instanceof Map) {
                m = (Map<String, Object>)v;
            } else {
                // throw new Exception("TypeError, nested Map expected");
                LOG.warn("Invalid nested config: {}, use default", key);
                return default_;
            }
        }
        if (v.getClass() != type) {
            // throw new Exception(String.format("TypeError, %s expected", type));
            LOG.warn("Invalid config type: {}, use default", key);
            return default_;
        }
        return v;
    }
}