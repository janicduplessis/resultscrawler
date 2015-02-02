package com.jduplessis.results.api;

import com.google.api.client.util.Key;

/**
 * Created by janic on 2015-02-02.
 */
public class Register {
    public static class Request {
        @Key
        public String email;

        @Key
        public String password;

        @Key
        public String firstName;

        @Key
        public String lastName;
    }
}
