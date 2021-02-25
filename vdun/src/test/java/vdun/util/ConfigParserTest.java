package vdun.util;

import org.junit.Test;
import static org.junit.Assert.*;

import java.util.HashMap;
import java.util.Map;

public class ConfigParserTest {
    @Test
    public void testGetBasicConfig() throws Exception {
        Map confMap = new HashMap<String, Object>() {{
            put("boolean", true);
            put("integer", 1);
            put("long", 1L);
            put("string", "");
        }};
        ConfigParser parser = new ConfigParser(confMap);
        assertEquals(Boolean.valueOf(true), parser.getBoolean("boolean", false));
        assertEquals(Boolean.valueOf(false), parser.getBoolean("x", false));
        assertEquals(Integer.valueOf(1), parser.getInteger("integer", 0));
        assertEquals(Integer.valueOf(0), parser.getInteger("x", 0));
        assertEquals(Long.valueOf(1), parser.getLong("long", 0L));
        assertEquals(Long.valueOf(0), parser.getLong("x", 0L));
        assertEquals("", parser.getString("string", "n"));
        assertEquals("n", parser.getString("x", "n"));
    }

    @Test
    public void testGetNestedConfig() throws Exception {
        Map confMap = new HashMap<String, Object>();
        Map v = new HashMap<String, Integer>();
        v.put("bar", 1);
        confMap.put("foo", v);
        ConfigParser parser = new ConfigParser(confMap);
        assertEquals(Integer.valueOf(1), parser.getInteger("foo.bar", 0));
        assertEquals(Integer.valueOf(0), parser.getInteger("foo.xxx", 0));
    }

    //@Test(expected=Exception.class)
    public void testInvalidNestedConfig() throws Exception {
        Map confMap = new HashMap<String, Object>();
        confMap.put("foo", "bar");
        ConfigParser parser = new ConfigParser(confMap);
        assertEquals(Integer.valueOf(1), parser.getInteger("foo.bar", 1));
    }

    //@Test(expected=Exception.class)
    public void testTypeNotMatchConfig() throws Exception {
        Map confMap = new HashMap<String, Object>();
        confMap.put("foo", "bar");
        ConfigParser parser = new ConfigParser(confMap);
        assertEquals(Integer.valueOf(1), parser.getInteger("foo", 1));
    }
}