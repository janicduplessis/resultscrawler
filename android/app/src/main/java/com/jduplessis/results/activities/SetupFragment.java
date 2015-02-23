package com.jduplessis.results.activities;

import android.animation.Animator;
import android.animation.AnimatorListenerAdapter;
import android.app.Activity;
import android.net.Uri;
import android.os.AsyncTask;
import android.os.Bundle;
import android.app.Fragment;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.Switch;
import android.widget.TextView;

import com.jduplessis.results.R;
import com.jduplessis.results.api.Client;
import com.jduplessis.results.api.CrawlerClass;
import com.jduplessis.results.api.CrawlerConfig;
import com.jduplessis.results.api.Results;

import org.w3c.dom.Text;

import java.io.IOException;
import java.util.List;

public class SetupFragment extends Fragment {
    private Client mClient = Client.getInstance();
    private AsyncTask mTask;

    private Switch mStatusSwitch;
    private TextView mCodeView;
    private TextView mNipView;
    private TextView mNotificationEmail;

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


            showProgress(false);
        }

        @Override
        protected void onCancelled() {
            mTask = null;
            showProgress(false);
        }
    }
}
