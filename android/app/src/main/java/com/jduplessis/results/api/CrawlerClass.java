package com.jduplessis.results.api;

import com.google.api.client.util.Key;

import java.util.ArrayList;

/**
 * Created by janic on 2015-02-23.
 */
public class CrawlerClass {
    public static class List extends ArrayList<CrawlerClass> {}

    @Key
    public String id;

    @Key
    public String name;

    @Key
    public String group;

    @Key
    public String year;
}
