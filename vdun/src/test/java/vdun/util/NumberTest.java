package vdun.util;

import org.junit.Test;
import static org.junit.Assert.*;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

public class NumberTest {
    @Test
    public void testStd() {
        List<Double> v0 = new ArrayList<Double>(Arrays.asList(1.0, 1.0));
        assertEquals(0.0, Number.std(v0), 1e-10);

        List<Double> v1 = new ArrayList<Double>(Arrays.asList(1.0, 2.0, 3.0, 4.0, 5.0));
        assertEquals(1.4142135623730951, Number.std(v1), 1e-10);
    }
}