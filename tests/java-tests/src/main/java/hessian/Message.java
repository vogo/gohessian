package hessian;

import java.io.Serializable;

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


    public static void main(String[] args) {

    }
}

