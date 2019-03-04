package hessian;

import com.caucho.hessian.io.Hessian2Input;
import com.caucho.hessian.io.Hessian2Output;
import com.caucho.hessian.io.SerializerFactory;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;

public class HessianTool {

    private static SerializerFactory serializerFactory = new SerializerFactory();

    public static byte[] serialize(Object obj) {
        ByteArrayOutputStream ops = new ByteArrayOutputStream();
        Hessian2Output out = new Hessian2Output(ops);
        out.setSerializerFactory(serializerFactory);

        try {
            out.writeObject(obj);
            out.close();
        } catch (IOException e) {
            throw new RuntimeException("hessian error", e);
        }

        byte[] bytes = ops.toByteArray();
        return bytes;
    }

    public static Object deserialize(byte[] bytes) {
        ByteArrayInputStream ips = new ByteArrayInputStream(bytes);
        Hessian2Input in = new Hessian2Input(ips);
        in.setSerializerFactory(serializerFactory);
        Object value = null;
        try {
            value = in.readObject();
            in.close();
        } catch (IOException e) {
            throw new RuntimeException("hessian error", e);
        }

        return value != null ? value : bytes;
    }
}
