package com.jduplessis.results.api;

import com.google.api.client.http.GenericUrl;
import com.google.api.client.http.HttpRequest;
import com.google.api.client.http.HttpRequestFactory;
import com.google.api.client.http.HttpRequestInitializer;
import com.google.api.client.http.HttpTransport;
import com.google.api.client.http.javanet.NetHttpTransport;
import com.google.api.client.http.json.JsonHttpContent;
import com.google.api.client.json.JsonFactory;
import com.google.api.client.json.JsonObjectParser;
import com.google.api.client.json.jackson.JacksonFactory;

import java.io.IOException;

/**
 * Created by Janic on 2015-01-25.
 */
public class Client {
    private static Client mInstance = new Client();
    private static final HttpTransport HTTP_TRANSPORT = new NetHttpTransport();
    private static final JsonFactory JSON_FACTORY = new JacksonFactory();

    private static final String URL_BASE = "http://results.jdupserver.com/api/v1/";
    private static final String URL_LOGIN = URL_BASE + "auth/login";
    private static final String URL_REGISTER = URL_BASE + "auth/register";
    private static final String URL_RESULTS = URL_BASE + "results/:session";

    private String mAuthToken;

    public static Client getInstance() {
        return mInstance;
    }

    private Client() {
    }

    public void setAuthToken(String token) {
        mAuthToken = token;
    }

    public Login.Response login(String email, String password) throws IOException {
        HttpRequestFactory requestFactory = getJSONRequestFactory();
        Login.Request requestData = new Login.Request();
        requestData.email = email;
        requestData.password = password;


        HttpRequest request = requestFactory.buildPostRequest(new GenericUrl(URL_LOGIN),
                new JsonHttpContent(JSON_FACTORY, requestData));

        return request.execute().parseAs(Login.Response.class);
    }

    public Login.Response register(String email, String password, String firstName, String lastName) throws IOException {
        HttpRequestFactory requestFactory = getJSONRequestFactory();
        Register.Request requestData = new Register.Request();
        requestData.email = email;
        requestData.password = password;
        requestData.firstName = firstName;
        requestData.lastName = lastName;

        HttpRequest request = requestFactory.buildPostRequest(new GenericUrl(URL_REGISTER),
                new JsonHttpContent(JSON_FACTORY, requestData));

        return request.execute().parseAs(Login.Response.class);
    }

    public Results getResults(String session) throws IOException {
        HttpRequestFactory requestFactory = getJSONRequestFactory();
        String url = URL_RESULTS.replace(":session", session);

        HttpRequest request = requestFactory.buildGetRequest(new GenericUrl(url));
        authenticateRequest(request);
        return request.execute().parseAs(Results.class);
    }

    private HttpRequestFactory getJSONRequestFactory() {
        return HTTP_TRANSPORT.createRequestFactory(new HttpRequestInitializer() {
            @Override
            public void initialize(HttpRequest request) throws IOException {
                request.setParser(new JsonObjectParser(JSON_FACTORY));
            }
        });
    }

    private void authenticateRequest(HttpRequest request) {
        if(mAuthToken != null) {
            request.getHeaders().set("X-Access-Token", mAuthToken);
        }
    }
}
