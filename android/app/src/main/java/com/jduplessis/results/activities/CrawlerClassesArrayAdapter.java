package com.jduplessis.results.activities;

import android.content.Context;
import android.text.TextUtils;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.ArrayAdapter;
import android.widget.TableLayout;
import android.widget.TableRow;
import android.widget.TextView;

import com.jduplessis.results.R;
import com.jduplessis.results.api.CrawlerClass;
import com.jduplessis.results.api.Results;
import com.jduplessis.results.api.Util;

import java.util.List;

/**
 * Created by Janic on 2015-02-27.
 */
public class CrawlerClassesArrayAdapter extends ArrayAdapter<CrawlerClass> {
    private final Context mContext;
    private final List<CrawlerClass> mValues;

    public CrawlerClassesArrayAdapter(Context context, List<CrawlerClass> values) {
        super(context, R.layout.layout_crawler_class, values);
        mContext = context;
        mValues = values;
    }

    @Override
    public View getView(int position, View convertView, ViewGroup parent) {
        LayoutInflater inflater = (LayoutInflater) mContext
                .getSystemService(Context.LAYOUT_INFLATER_SERVICE);
        View rowView = inflater.inflate(R.layout.layout_crawler_class, parent, false);

        TextView classNameView = (TextView)rowView.findViewById(R.id.class_name_txt);
        TextView sessionView = (TextView)rowView.findViewById(R.id.session_txt);

        CrawlerClass c = mValues.get(position);
        classNameView.setText(c.name + " - " + c.group);
        sessionView.setText(Util.formatSession(c.year));

        return rowView;
    }
}
