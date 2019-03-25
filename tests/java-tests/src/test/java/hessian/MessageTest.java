// Copyright 2018-2019 vogo.
// Author: wongoo
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

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
