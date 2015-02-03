package com.jduplessis.results.activities;

import android.content.Context;
import android.text.TextUtils;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.ArrayAdapter;
import android.widget.ListView;
import android.widget.TableLayout;
import android.widget.TableRow;
import android.widget.TextView;

import com.jduplessis.results.R;
import com.jduplessis.results.api.Results;

import java.util.List;

/**
 * Created by Janic on 2015-01-25.
 */
public class ResultsArrayAdapter extends ArrayAdapter<Results.Class> {
    private final Context mContext;
    private final List<Results.Class> mValues;

    public ResultsArrayAdapter(Context context, List<Results.Class> values) {
        super(context, R.layout.layout_class_card, values);
        mContext = context;
        mValues = values;
    }

    @Override
    public View getView(int position, View convertView, ViewGroup parent) {
        LayoutInflater inflater = (LayoutInflater) mContext
                .getSystemService(Context.LAYOUT_INFLATER_SERVICE);
        View rowView = inflater.inflate(R.layout.layout_class_card, parent, false);

        TextView classNameView = (TextView)rowView.findViewById(R.id.txtClassName);
        TextView totalResultView = (TextView)rowView.findViewById(R.id.txtTotalResult);
        TextView totalAverageView = (TextView)rowView.findViewById(R.id.txtTotalAverage);
        TableLayout resultsList = (TableLayout)rowView.findViewById(R.id.tableResults);
        View finalRowView = rowView.findViewById(R.id.rowFinalGrade);
        TextView finalGradeView = (TextView)rowView.findViewById(R.id.txtFinalGrade);

        Results.Class c = mValues.get(position);
        classNameView.setText(c.name);
        totalResultView.setText(c.total.result);
        totalAverageView.setText(c.total.average);
        if(TextUtils.isEmpty(c.finalGrade)) {
            finalRowView.setVisibility(View.GONE);
        } else {
            finalGradeView.setText(c.finalGrade);
        }

        for(int i = 0; i < c.results.size(); i++) {
            Results.Result res = c.results.get(i);
            TableRow row = new TableRow(mContext);
            TextView txtName = new TextView(mContext);
            txtName.setLayoutParams(new TableRow.LayoutParams(ViewGroup.LayoutParams.WRAP_CONTENT, ViewGroup.LayoutParams.WRAP_CONTENT, 1f));
            txtName.setText(res.name);
            row.addView(txtName);

            TextView txtResult = new TextView(mContext);
            txtResult.setLayoutParams(new TableRow.LayoutParams(ViewGroup.LayoutParams.WRAP_CONTENT, ViewGroup.LayoutParams.WRAP_CONTENT, 1f));
            txtResult.setText(res.normal.result);
            row.addView(txtResult);

            TextView txtAverage = new TextView(mContext);
            txtAverage.setLayoutParams(new TableRow.LayoutParams(ViewGroup.LayoutParams.WRAP_CONTENT, ViewGroup.LayoutParams.WRAP_CONTENT, 1f));
            txtAverage.setText(res.normal.average);
            row.addView(txtAverage);

            resultsList.addView(row, i+1);
        }

        return rowView;
    }
}
