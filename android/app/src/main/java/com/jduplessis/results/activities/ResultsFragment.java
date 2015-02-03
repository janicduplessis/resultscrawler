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
import android.view.View;
import android.view.ViewGroup;

import com.jduplessis.results.R;
import com.jduplessis.results.api.Client;
import com.jduplessis.results.api.Login;
import com.jduplessis.results.api.Results;

import java.io.IOException;
import java.util.ArrayList;

public class ResultsFragment extends ListFragment {
    private Client mClient;
    private ResultsTask mResultsTask = null;
    ResultsArrayAdapter mAdapter;

    public ResultsFragment() {
        // Required empty public constructor
    }

    @Override
    public void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        mAdapter = new ResultsArrayAdapter(getActivity(), new ArrayList<Results.Class>());
        setListAdapter(mAdapter);

        mClient = Client.getInstance();

        ResultsTask task = new ResultsTask("20143");
        task.execute();
    }

    @Override
    public View onCreateView(LayoutInflater inflater, ViewGroup container,
                             Bundle savedInstanceState) {
        // Inflate the layout for this fragment
        return inflater.inflate(R.layout.fragment_results, container, false);
    }

    @Override
    public void onAttach(Activity activity) {
        super.onAttach(activity);
    }

    @Override
    public void onDetach() {
        super.onDetach();
    }

    public void onSessionChange(String session) {

    }

    public void onRefresh() {

    }

    public class ResultsTask extends AsyncTask<Void, Void, Results> {

        private final String mSession;

        ResultsTask(String session) {
            mSession = session;
        }

        @Override
        protected Results doInBackground(Void... params) {
            Client client = Client.getInstance();
            Results response = null;
            try {
                response = client.getResults(mSession);
            } catch (IOException e) {
                e.printStackTrace();
                return null;
            }

            return response;
        }

        @Override
        protected void onPostExecute(final Results results) {
            mResultsTask = null;
            mAdapter.addAll(results.classes);

            //showProgress(false);
        }

        @Override
        protected void onCancelled() {
            mResultsTask = null;
            //showProgress(false);
        }
    }
}
