package hessian;

import org.junit.Assert;
import org.junit.Test;

import java.util.Random;

public class ArrayNameTest {

    @Test
    public void testArrayName() {
        ArrayHolder holder = new ArrayHolder();
        byte[] bytes = new byte[10];
        new Random().nextBytes(bytes);
        holder.setBytes(bytes);

        holder.setBooleans(new boolean[]{true, false, true});
        holder.setDoubles(new double[]{12.34, 56.78});
        holder.setFloats(new float[]{33.33f, 44.44f});
        holder.setInts(new int[]{111, 222});
        holder.setThreeDimeInts(new int[][][]{
                new int[][]{
                        new int[]{111, 222}
                },
                new int[][]{
                        new int[]{333, 444}
                },
                new int[][]{
                        new int[]{555, 666}
                }});
        holder.setLongs(new long[]{333l, 555l});
        holder.setStrings(new String[]{"hello", "world"});
        holder.setVos(new TraceVo[]{new TraceVo("k1", "v1"), new TraceVo("k2", "v2")});

        byte[] out = HessianTool.serialize(holder);
        String outString = new String(out);
        System.out.println(outString);

        Assert.assertTrue(outString.contains("[boolean"));
        Assert.assertTrue(outString.contains("[[[int"));
        Assert.assertTrue(outString.contains("[[int"));
        Assert.assertTrue(outString.contains("[int"));
        Assert.assertTrue(outString.contains("[long"));
        Assert.assertTrue(outString.contains("[double"));
        Assert.assertTrue(outString.contains("[float"));
        Assert.assertTrue(outString.contains("[string"));
        Assert.assertTrue(outString.contains("[" + TraceVo.class.getName()));
    }
}
