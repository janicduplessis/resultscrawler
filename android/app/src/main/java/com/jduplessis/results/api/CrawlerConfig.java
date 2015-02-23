package com.jduplessis.results.api;

import com.google.api.client.util.Key;

/**
 * Created by janic on 2015-02-23.
 */
public class CrawlerConfig {
    @Key
    public boolean status;

    @Key
    public String code;

    @Key
    public String nip;

    @Key
    public String notificationEmail;
}
