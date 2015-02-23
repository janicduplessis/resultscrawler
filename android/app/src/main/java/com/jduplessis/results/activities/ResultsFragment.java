package com.jduplessis.results.activities;

import android.accounts.Account;
import android.accounts.AccountManager;
import android.app.Activity;
import android.app.ListFragment;
import android.content.Intent;
import android.net.Uri;
import android.os.AsyncTask;
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

import com.jduplessis.results.R;
import com.jduplessis.results.api.Client;
import com.jduplessis.results.api.Login;
import com.jduplessis.results.api.Results;
import com.jduplessis.results.api.Util;

import java.io.IOException;
import java.util.ArrayList;

public class ResultsFragment extends ListFragment {
    private Client mClient;
    private ResultsTask mResultsTask = null;
    ResultsArrayAdapter mAdapter;
    String mCurrentSession;
    ArrayList<String> mSessions;

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

        updateResults();
    }

    @Override
    public View onCreateView(LayoutInflater inflater, ViewGroup container,
                             Bundle savedInstanceState) {
        // Inflate the layout for this fragment
        View view = inflater.inflate(R.layout.fragment_results, container, false);

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
                mCurrentSession = mSessions.get(position);
                updateResults();
            }

            @Override
            public void onNothingSelected(AdapterView<?> parent) {

            }
        });

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
        RefreshTask task = new RefreshTask();
        task.execute();
    }

    private void updateResults() {
        ResultsTask task = new ResultsTask(mCurrentSession);
        task.execute();
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
            mResultsTask = null;
            mAdapter.clear();
            mAdapter.addAll(results.classes);

            //showProgress(false);
        }

        @Override
        protected void onCancelled() {
            mResultsTask = null;
            //showProgress(false);
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
            mResultsTask = null;
            //showProgress(false);
            updateResults();
        }

        @Override
        protected void onCancelled() {
            mResultsTask = null;
            //showProgress(false);
        }
    }
}
