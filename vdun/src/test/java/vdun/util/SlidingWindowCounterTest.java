package vdun.util;

import static org.junit.Assert.*;

import java.util.List;
import java.util.Map.Entry;

import org.junit.Test;

public class SlidingWindowCounterTest {
    @Test(expected = IllegalArgumentException.class)
    public void testInvalidWindowLength() {
        new SlidingWindowCounter<String>(1);
    }

    @Test
    public void testCounter() {
        SlidingWindowCounter<String> counter =  new SlidingWindowCounter<String>(2);

        counter.increase("#1", 1);
        assertEquals(counter.getCounts().get("#1"), Integer.valueOf(1));
        assertEquals(counter.getCountsThenAdvanceWindow().get("#1"), Integer.valueOf(1));

        counter.increase("#1", Integer.valueOf(1));
        assertEquals(counter.getCounts().get("#1"), Integer.valueOf(2));
        assertEquals(counter.getCountsThenAdvanceWindow().get("#1"), Integer.valueOf(2));
        assertEquals(counter.getCounts().get("#1"), Integer.valueOf(1));

        counter.increase("#1", Integer.valueOf(1));
        assertEquals(counter.getCounts().get("#1"), Integer.valueOf(2));
        assertEquals(counter.getCountsThenAdvanceWindow().get("#1"), Integer.valueOf(2));

        assertEquals(counter.getCountsThenAdvanceWindow().get("#1"), Integer.valueOf(1));        
        assertEquals(counter.getCountsThenAdvanceWindow().get("#1"), null);

        counter.increase("#1", Integer.valueOf(1));
        assertEquals(counter.getCounts().get("#1"), Integer.valueOf(1));
    }

    @Test
    public void testGetMostCommonWithLessCounter() {
        SlidingWindowCounter<String> counter = new SlidingWindowCounter<String>(2);
        assertEquals(counter.getMostCommon(5).size(), 0);

        counter.increase("#1", 1);
        counter.increase("#2", 2);
        counter.increase("#3", 3);
        List<Entry<String, Integer>> top5 = counter.getMostCommon(5);
        assertEquals(top5.size(), 3);
        assertEquals(top5.get(0).getKey(), "#3");
        assertEquals(top5.get(0).getValue(), Integer.valueOf(3));
        assertEquals(top5.get(1).getKey(), "#2");
        assertEquals(top5.get(1).getValue(), Integer.valueOf(2));
        assertEquals(top5.get(2).getKey(), "#1");
        assertEquals(top5.get(2).getValue(), Integer.valueOf(1));        
    }

    @Test
    public void testGetMostCommonWithMoreCounter() {
        SlidingWindowCounter<String> counter = new SlidingWindowCounter<String>(2);
        counter.increase("#1", 1);
        counter.increase("#2", 1);
        counter.increase("#3", 1);
        assertEquals(counter.getMostCommon(2).size(), 2);
        assertEquals(counter.getMostCommon().size(), 3);
    }

    @Test
    public void testAutoRemoveZeroCounter() {
        SlidingWindowCounter<String> counter = new SlidingWindowCounter<String>(2);
        counter.increase("#1", 1);
        assertEquals(counter.size(), 1);
        counter.advanceWindow();
        assertEquals(counter.size(), 1);
        counter.advanceWindow();
        assertEquals(counter.size(), 0);
    }
}