package hessian;

import org.apache.commons.codec.binary.Base64;
import org.junit.Assert;
import org.junit.Test;

import java.util.ArrayList;
import java.util.List;

public class MessageTest {

    @Test
    public void testEncode() {
        TraceVo v1 = new TraceVo();
        TraceVo v2 = new TraceVo();
        v1.setKey("k1");
        v1.setValue("v1");
        v2.setKey("k2");
        v2.setValue("v2");

        TraceData<TraceVo> d1 = new TraceData<>();
        TraceData<TraceVo> d2 = new TraceData<>();
        d1.setSeq(123456);
        d1.setData(v1);
        d2.setSeq(123457);
        d2.setData(v2);

        List<TraceData<TraceVo>> list = new ArrayList<>();
        list.add(d1);
        list.add(d2);

        Message<List<TraceData<TraceVo>>> message = new Message<>();
        message.setTitle("m1");
        message.setMsg(list);

        byte[] bytes = HessianTool.serialize(message);
        Assert.assertNotNull(bytes);

        String base64 = Base64.encodeBase64String(bytes);
        System.out.println(base64);
    }
}
