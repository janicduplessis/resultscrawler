package com.jduplessis.results.activities;

import android.animation.Animator;
import android.animation.AnimatorListenerAdapter;
import android.app.Activity;
import android.content.Context;
import android.net.Uri;
import android.os.AsyncTask;
import android.os.Bundle;
import android.app.Fragment;
import android.view.LayoutInflater;
import android.view.MotionEvent;
import android.view.View;
import android.view.ViewGroup;
import android.widget.ArrayAdapter;
import android.widget.CompoundButton;
import android.widget.LinearLayout;
import android.widget.ListView;
import android.widget.Switch;
import android.widget.TextView;

import com.jduplessis.results.R;
import com.jduplessis.results.api.Client;
import com.jduplessis.results.api.CrawlerClass;
import com.jduplessis.results.api.CrawlerConfig;
import com.jduplessis.results.api.Results;
import com.jduplessis.results.api.Util;

import org.w3c.dom.Text;

import java.io.IOException;
import java.util.ArrayList;
import java.util.List;

public class SetupFragment extends Fragment {
    private Client mClient = Client.getInstance();
    private AsyncTask mTask;

    private Switch mStatusSwitch;
    private TextView mCodeView;
    private TextView mNipView;
    private TextView mNotificationEmail;
    private LinearLayout mClassesView;

    private boolean mProgressVisible = false;
    private View mProgressView;
    private View mContentView;

    public SetupFragment() {
        // Required empty public constructor
    }

    @Override
    public void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
    }

    @Override
    public View onCreateView(LayoutInflater inflater, ViewGroup container,
                             Bundle savedInstanceState) {
        // Inflate the layout for this fragment
        View view = inflater.inflate(R.layout.fragment_setup, container, false);

        mStatusSwitch = (Switch)view.findViewById(R.id.state_switch);
        mCodeView = (TextView)view.findViewById(R.id.code_text);
        mNipView = (TextView)view.findViewById(R.id.nip_text);
        mNotificationEmail = (TextView)view.findViewById(R.id.notification_email_text);
        mProgressView = view.findViewById(R.id.loading_progress);
        mContentView = view.findViewById(R.id.main_content);
        mClassesView = (LinearLayout)view.findViewById(R.id.list_classes);

        View.OnFocusChangeListener controlChangedListener = new View.OnFocusChangeListener() {
            @Override
            public void onFocusChange(View v, boolean hasFocus) {
                if(!hasFocus) {
                    saveCrawlerConfig();
                }
            }
        };
        mCodeView.setOnFocusChangeListener(controlChangedListener);
        mNipView.setOnFocusChangeListener(controlChangedListener);
        mNotificationEmail.setOnFocusChangeListener(controlChangedListener);
        mStatusSwitch.setOnCheckedChangeListener(new CompoundButton.OnCheckedChangeListener() {
            @Override
            public void onCheckedChanged(CompoundButton buttonView, boolean isChecked) {
                saveCrawlerConfig();
            }
        });

        updateConfig();

        return view;
    }

    @Override
    public void onAttach(Activity activity) {
        super.onAttach(activity);
    }

    @Override
    public void onDetach() {
        super.onDetach();
    }

    private void updateConfig() {
        if(mTask != null) {
            return;
        }
        showProgress(true);
        GetCrawlerConfigTask task = new GetCrawlerConfigTask();
        mTask = task;
        GetCrawlerClassesTask task2 = new GetCrawlerClassesTask();
        task.execute();
        task2.execute();
    }

    private void saveCrawlerConfig() {
        SaveCrawlerConfigTask task = new SaveCrawlerConfigTask();
        task.execute();
    }

    private void showProgress(final boolean show) {
        int shortAnimTime = getResources().getInteger(android.R.integer.config_shortAnimTime);
        if(show && mProgressVisible || !show && !mProgressVisible) {
            return;
        }
        mProgressVisible = show;
        mContentView.setVisibility(show ? View.GONE : View.VISIBLE);
        mContentView.animate().setDuration(shortAnimTime).alpha(
                show ? 0 : 1).setListener(new AnimatorListenerAdapter() {
            @Override
            public void onAnimationEnd(Animator animation) {
                mContentView.setVisibility(show ? View.GONE : View.VISIBLE);
            }
        });

        mProgressView.setVisibility(show ? View.VISIBLE : View.GONE);
        mProgressView.animate().setDuration(shortAnimTime).alpha(
                show ? 1 : 0).setListener(new AnimatorListenerAdapter() {
            @Override
            public void onAnimationEnd(Animator animation) {
                mProgressView.setVisibility(show ? View.VISIBLE : View.GONE);
            }
        });
    }

    public class GetCrawlerConfigTask extends AsyncTask<Void, Void, CrawlerConfig> {

        @Override
        protected CrawlerConfig doInBackground(Void... params) {
            CrawlerConfig response = null;
            try {
                response = mClient.getCrawlerConfig();
            } catch (IOException e) {
                e.printStackTrace();
                return null;
            }

            return response;
        }

        @Override
        protected void onPostExecute(final CrawlerConfig config) {
            mTask = null;

            mStatusSwitch.setChecked(config.status);
            mCodeView.setText(config.code);
            mNipView.setText(config.nip);
            mNotificationEmail.setText(config.notificationEmail);

            showProgress(false);
        }

        @Override
        protected void onCancelled() {
            mTask = null;
            showProgress(false);
        }
    }

    public class SaveCrawlerConfigTask extends AsyncTask<Void, Void, Boolean> {
        @Override
        protected Boolean doInBackground(Void... params) {
            try {
                // TODO: check if save failed.
                mClient.saveCrawlerConfig(mStatusSwitch.isChecked(), mCodeView.getText().toString(),
                        mNipView.getText().toString(), mNotificationEmail.getText().toString());
            } catch (IOException e) {
                e.printStackTrace();
                return false;
            }

            return true;
        }
    }

    public class GetCrawlerClassesTask extends AsyncTask<Void, Void, List<CrawlerClass>> {

        @Override
        protected List<CrawlerClass> doInBackground(Void... params) {
            List<CrawlerClass> response = null;
            try {
                response = mClient.getConfigClasses();
            } catch (IOException e) {
                e.printStackTrace();
                return null;
            }

            return response;
        }

        @Override
        protected void onPostExecute(final List<CrawlerClass> classes) {
            mTask = null;

            LayoutInflater inflater = (LayoutInflater) getActivity()
                    .getSystemService(Context.LAYOUT_INFLATER_SERVICE);

            mClassesView.removeAllViews();
            for(int i = 0; i < classes.size(); i++) {
                CrawlerClass c = classes.get(i);
                View rowView = inflater.inflate(R.layout.layout_crawler_class, mClassesView, false);

                TextView classNameView = (TextView)rowView.findViewById(R.id.class_name_txt);
                TextView sessionView = (TextView)rowView.findViewById(R.id.session_txt);

                classNameView.setText(c.name + " - " + c.group);
                sessionView.setText(Util.formatSession(c.year));

                if(i == classes.size() - 1) {
                    View divider = rowView.findViewById(R.id.divider);
                    divider.setVisibility(View.GONE);
                }

                mClassesView.addView(rowView);
            }
            showProgress(false);
        }

        @Override
        protected void onCancelled() {
            mTask = null;
            showProgress(false);
        }
    }
}
