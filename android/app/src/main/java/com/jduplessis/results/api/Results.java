package com.jduplessis.results.api;

import com.google.api.client.json.GenericJson;
import com.google.api.client.util.DateTime;
import com.google.api.client.util.Key;

import java.util.ArrayList;
import java.util.List;

/**
 * Created by Janic on 2015-01-25.
 */
public class Results {
    @Key
    public DateTime lastUpdate;

    @Key
    public List<Class> classes;

    public static class Class extends GenericJson {
        @Key
        public String id;

        @Key
        public String name;

        @Key
        public String group;

        @Key
        public String year;

        @Key
        public List<Result> results;

        @Key
        public ResultInfo total;

        @Key("final")
        public String finalGrade;
    }

    public static class Result extends GenericJson {
        @Key
        public String name;

        @Key
        public ResultInfo normal;

        @Key
        public ResultInfo weighted;
    }

    public static class ResultInfo extends GenericJson {
        @Key
        public String result;

        @Key
        public String average;

        @Key
        public String standardDev;
    }
}
