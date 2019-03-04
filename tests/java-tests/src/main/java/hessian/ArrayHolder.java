package hessian;

import java.io.Serializable;

public class ArrayHolder implements Serializable {
    public byte[] bytes;
    public int[] ints;
    public int[][][] threeDimeInts;
    public String[] strings;
    public double[] doubles;
    public float[] floats;
    public long[] longs;

    public int[][][] getThreeDimeInts() {
        return threeDimeInts;
    }

    public void setThreeDimeInts(int[][][] threeDimeInts) {
        this.threeDimeInts = threeDimeInts;
    }

    public boolean[] getBooleans() {
        return booleans;
    }

    public void setBooleans(boolean[] booleans) {
        this.booleans = booleans;
    }

    public boolean[] booleans;

    public long[] getLongs() {
        return longs;
    }

    public void setLongs(long[] longs) {
        this.longs = longs;
    }

    public byte[] getBytes() {
        return bytes;
    }

    public void setBytes(byte[] bytes) {
        this.bytes = bytes;
    }

    public int[] getInts() {
        return ints;
    }

    public void setInts(int[] ints) {
        this.ints = ints;
    }

    public String[] getStrings() {
        return strings;
    }

    public void setStrings(String[] strings) {
        this.strings = strings;
    }

    public double[] getDoubles() {
        return doubles;
    }

    public void setDoubles(double[] doubles) {
        this.doubles = doubles;
    }

    public float[] getFloats() {
        return floats;
    }

    public void setFloats(float[] floats) {
        this.floats = floats;
    }

    public TraceVo[] getVos() {
        return vos;
    }

    public void setVos(TraceVo[] vos) {
        this.vos = vos;
    }

    public TraceVo[] vos;
}
