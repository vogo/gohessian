package hessian;

import org.junit.Assert;
import org.junit.Test;

import java.util.Calendar;
import java.util.Date;

public class DateTest {

    @Test
    public void testDateFormat() {
        byte[] bytes = HessianTool.serialize(new Date());
        Assert.assertEquals(9, bytes.length);
        for (byte b : bytes) {
            System.out.printf("%x ", b);
        }
        System.out.println();

        Calendar calendar = Calendar.getInstance();
        calendar.set(Calendar.MILLISECOND, 0);

        bytes = HessianTool.serialize(calendar.getTime());
        Assert.assertEquals(5, bytes.length);
        for (byte b : bytes) {
            System.out.printf("%x ", b);
        }
        System.out.println();
    }
}
