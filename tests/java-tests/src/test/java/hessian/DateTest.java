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
