package tests.jobj;

import com.caucho.hessian.io.Hessian2Output;
import com.caucho.hessian.io.SerializerFactory;
import org.apache.commons.codec.binary.Base64;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.Serializable;
import java.util.ArrayList;
import java.util.List;

public class Message<T> implements Serializable {

    public String title;
    public T msg;

    public String getTitle() {
        return title;
    }

    public void setTitle(String title) {
        this.title = title;
    }

    public T getMsg() {
        return msg;
    }

    public void setMsg(T msg) {
        this.msg = msg;
    }

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

    public static void main(String[] args) {
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

        byte[] bytes = serialize(message);
        String base64 = Base64.encodeBase64String(bytes);
        System.out.println(base64);
    }
}

class TraceData<E> implements Serializable {
    public int seq;
    public E data;

    public int getSeq() {
        return seq;
    }

    public void setSeq(int seq) {
        this.seq = seq;
    }

    public E getData() {
        return data;
    }

    public void setData(E data) {
        this.data = data;
    }
}

class TraceVo implements Serializable {
    public String key;
    public String value;

    public String getKey() {
        return key;
    }

    public void setKey(String key) {
        this.key = key;
    }

    public String getValue() {
        return value;
    }

    public void setValue(String value) {
        this.value = value;
    }
}
