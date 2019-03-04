package hessian;

import java.io.Serializable;

public class TraceVo<E> implements Serializable {

    public String key;
    public String value;

    public TraceVo() {
    }

    public TraceVo(String key, String value) {
        this.key = key;
        this.value = value;
    }

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