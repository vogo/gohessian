package hessian;

import java.io.Serializable;

public class TraceData<E> implements Serializable {

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
