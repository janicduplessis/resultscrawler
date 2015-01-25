package com.jduplessis.results.api;

import com.google.api.client.json.GenericJson;
import com.google.api.client.util.Key;

import java.util.ArrayList;

/**
 * Created by Janic on 2015-01-25.
 */
public class Results {
    @Key
    String lastUpdate;

    @Key
    ArrayList<Class> classes;

    public class Class extends GenericJson {
        @Key
        String id;

        @Key
        String name;

        @Key
        String group;

        @Key
        String year;

        @Key
        ArrayList<Result> results;

        @Key
        ResultInfo total;

        @Key("final")
        String finalGrade;
    }

    public class Result extends GenericJson {
        @Key
        String name;

        @Key
        ResultInfo normal;

        @Key
        ResultInfo weighted;
    }

    public class ResultInfo extends GenericJson {
        @Key
        String result;

        @Key
        String average;

        @Key
        String standardDev;
    }
}
