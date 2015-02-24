package com.jduplessis.results.activities;

import android.accounts.Account;
import android.accounts.AccountManager;
import android.animation.Animator;
import android.animation.AnimatorListenerAdapter;
import android.app.Activity;
import android.app.ListFragment;
import android.content.Intent;
import android.net.Uri;
import android.os.AsyncTask;
import android.os.Build;
import android.os.Bundle;
import android.app.Fragment;
import android.view.LayoutInflater;
import android.view.Menu;
import android.view.MenuInflater;
import android.view.MenuItem;
import android.view.View;
import android.view.ViewGroup;
import android.widget.AdapterView;
import android.widget.ArrayAdapter;
import android.widget.Spinner;
import android.widget.TextView;

import com.jduplessis.results.R;
import com.jduplessis.results.api.Client;
import com.jduplessis.results.api.Login;
import com.jduplessis.results.api.Results;
import com.jduplessis.results.api.Util;

import java.io.IOException;
import java.text.Format;
import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.Calendar;
import java.util.Date;

public class ResultsFragment extends ListFragment {
    ResultsArrayAdapter mAdapter;
    String mCurrentSession;
    ArrayList<String> mSessions;
    private Client mClient;
    private AsyncTask mTask = null;

    private View mProgressView;
    private boolean mProgressVisible = false;
    private View mContentView;
    private TextView mLastUpdateView;

    public ResultsFragment() {
        // Required empty public constructor
    }

    @Override
    public void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        setHasOptionsMenu(true);

        mAdapter = new ResultsArrayAdapter(getActivity(), new ArrayList<Results.Class>());
        setListAdapter(mAdapter);

        mClient = Client.getInstance();

        mCurrentSession = Util.getCurrentSession();
        mSessions = Util.getRecentSessions(6);
    }

    @Override
    public View onCreateView(LayoutInflater inflater, ViewGroup container,
                             Bundle savedInstanceState) {
        // Inflate the layout for this fragment
        View view = inflater.inflate(R.layout.fragment_results, container, false);

        mProgressView = view.findViewById(R.id.loading_progress);
        mContentView = view.findViewById(R.id.main_content);
        mLastUpdateView = (TextView)view.findViewById(R.id.last_update_text);

        ArrayList<String> sessionNames = new ArrayList<>(mSessions.size());
        for(String session: mSessions) {
            sessionNames.add(Util.formatSession(session));
        }

        Spinner spinner = (Spinner) view.findViewById(R.id.session_spinner);
        ArrayAdapter<String> adapter = new ArrayAdapter<>(getActivity(), android.R.layout.simple_spinner_item, sessionNames);
        adapter.setDropDownViewResource(android.R.layout.simple_spinner_dropdown_item);
        spinner.setAdapter(adapter);

        spinner.setOnItemSelectedListener(new AdapterView.OnItemSelectedListener() {
            @Override
            public void onItemSelected(AdapterView<?> parent, View view, int position, long id) {
                if(!mCurrentSession.equals(mSessions.get(position))) {
                    mCurrentSession = mSessions.get(position);
                    updateResults();
                }
            }

            @Override
            public void onNothingSelected(AdapterView<?> parent) {

            }
        });

        updateResults();

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

    @Override
    public void onCreateOptionsMenu(Menu menu, MenuInflater inflater) {
        inflater.inflate(R.menu.results, menu);
    }

    @Override
    public boolean onOptionsItemSelected(MenuItem item) {
        if(item.getItemId() == R.id.refresh) {
            onRefresh();
            return true;
        }

        return super.onOptionsItemSelected(item);
    }

    public void onSessionChange(String session) {

    }

    public void onRefresh() {
        if(mTask != null) {
            return;
        }
        showProgress(true);
        RefreshTask task = new RefreshTask();
        mTask = task;
        task.execute();
    }

    private void updateResults() {
        if(mTask != null) {
            return;
        }
        showProgress(true);
        ResultsTask task = new ResultsTask(mCurrentSession);
        mTask = task;
        task.execute();
    }

    public void showProgress(final boolean show) {
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

    public class ResultsTask extends AsyncTask<Void, Void, Results> {

        private final String mSession;

        ResultsTask(String session) {
            mSession = session;
        }

        @Override
        protected Results doInBackground(Void... params) {
            Results response = null;
            try {
                response = mClient.getResults(mSession);
            } catch (IOException e) {
                e.printStackTrace();
                return null;
            }

            return response;
        }

        @Override
        protected void onPostExecute(final Results results) {
            mTask = null;
            mAdapter.clear();
            mAdapter.addAll(results.classes);
            Format formatter = SimpleDateFormat.getDateTimeInstance();
            Date date = new Date(results.lastUpdate.getValue());
            mLastUpdateView.setText(formatter.format(date));

            showProgress(false);
        }

        @Override
        protected void onCancelled() {
            mTask = null;
            showProgress(false);
        }
    }

    public class RefreshTask extends AsyncTask<Void, Void, Boolean> {

        @Override
        protected Boolean doInBackground(Void... params) {
            try {
                mClient.refresh();
                return true;
            } catch (IOException e) {
                e.printStackTrace();
                return false;
            }
        }

        @Override
        protected void onPostExecute(final Boolean ok) {
            mTask = null;
            updateResults();
        }

        @Override
        protected void onCancelled() {
            mTask = null;
            showProgress(false);
        }
    }
}
