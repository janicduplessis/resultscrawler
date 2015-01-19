package com.jduplessis.results.models;

import com.google.api.client.util.Key;

/**
 * Created by Janic on 2015-01-17.
 */
public class Login {

    public static final int CODE_OK = 0;
    public static final int CODE_INVALID = 1;

    public static class Request {
        @Key
        public String email;

        @Key
        public String password;
    }

    public static class Response {
        @Key
        public int status;

        @Key
        public String authToken;

        @Key
        public User user;
    }
}
