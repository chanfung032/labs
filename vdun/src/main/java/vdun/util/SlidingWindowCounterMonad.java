package vdun.util;

public class SlidingWindowCounterMonad {
    private int[] slot;
    private int headSlot;
    private int windowLength;

    public SlidingWindowCounterMonad(int windowLength) {
        slot = new int[windowLength + 1];
        headSlot = 0;
        this.windowLength = windowLength;
    }

    public void increase(int n) {
        slot[headSlot] += n;
    }

    public int getCount() {
        return slot[windowLength] + slot[headSlot];
    }

    public int getCountThenAdvance() {
        slot[windowLength] += slot[headSlot];
        int result = slot[windowLength];
        headSlot = (headSlot + 1) % windowLength;
        slot[windowLength] -= slot[headSlot];
        slot[headSlot] = 0;
        return result;
    }
}