package vdun.util;

import org.junit.Test;
import static org.junit.Assert.*;

import java.util.List;
import java.util.Map.Entry;

public class CounterTest {
    @Test
    public void testCounter() {
        Counter<String> counter = new Counter<String>();
        counter.increase("#1", 1);
        assertEquals(counter.get("#1"), Integer.valueOf(1));
        counter.increase("#1");
        assertEquals(counter.get("#1"), Integer.valueOf(2));
        counter.increase("#2");
        List<Entry<String, Integer>> most = counter.getMostCommon();
        assertEquals(most.size(), 2);
        assertEquals(most.get(0).getKey(), "#1");
        assertEquals(most.get(1).getKey(), "#2");
    }
}