package com.jduplessis.results.api;

import java.util.ArrayList;
import java.util.Calendar;
import java.util.Date;
import java.util.GregorianCalendar;

/**
 * Created by Janic on 2015-02-22.
 */
public class Util {
    public static String formatSession(String session) {
        String result = null;
        char time = session.charAt(4);
        switch (time) {
            case '1':
                result = "Winter";
                break;
            case '2':
                result = "Summer";
                break;
            case '3':
                result = "Fall";
                break;
            default:
                // Invalid format
                return "";
        }

        return result + " " + session.substring(0, 4);
    }

    public static String getCurrentSession() {
        Date curDate = new Date();
        Calendar calendar = new GregorianCalendar();
        calendar.setTime(curDate);
        int curMonth = calendar.get(Calendar.MONTH);
        int curYear = calendar.get(Calendar.YEAR);
        String timeOfYear;
        if(curMonth >= Calendar.JANUARY && curMonth < Calendar.JUNE) {
            timeOfYear = "1";
        } else if(curMonth >= Calendar.JUNE && curMonth < Calendar.SEPTEMBER) {
            timeOfYear = "2";
        } else {
            timeOfYear = "3";
        }
        return String.valueOf(curYear) + timeOfYear;
    }

    public static ArrayList<String> getRecentSessions(int amount) {
        String curSession = getCurrentSession();
        ArrayList<String> result = new ArrayList<String>(amount);
        for(int i = 0; i < amount; i++) {
            result.add(curSession);
            int timeOfYear = Integer.parseInt(curSession.substring(4, 5));
            if(timeOfYear > 1) {
                timeOfYear--;
                curSession = curSession.substring(0, 4) + timeOfYear;
            } else {
                int year = Integer.parseInt(curSession.substring(0, 4));
                year--;
                curSession = year + "3";
            }
        }
        return result;
    }
}
