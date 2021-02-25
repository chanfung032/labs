package vdun.util;

import org.junit.Test;
import static org.junit.Assert.*;

public class SlidingWindowCounterMonadTest {
    @Test
    public void testCounter() {
        SlidingWindowCounterMonad counter = new SlidingWindowCounterMonad(2);

        counter.increase(1);
        assertEquals(counter.getCount(), 1);
        assertEquals(counter.getCountThenAdvance(), 1);
        counter.increase(1);
        assertEquals(counter.getCountThenAdvance(), 2);
        assertEquals(counter.getCount(), 1);
        assertEquals(counter.getCountThenAdvance(), 1);
        assertEquals(counter.getCountThenAdvance(), 0);
    }
}