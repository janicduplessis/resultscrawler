package com.jduplessis.results.models;

import com.google.api.client.json.GenericJson;
import com.google.api.client.util.Key;

/**
 * Created by Janic on 2015-01-15.
 */
public class User extends GenericJson {
    @Key
    public String email;

    @Key
    public String firstName;

    @Key
    public String lastName;

}
