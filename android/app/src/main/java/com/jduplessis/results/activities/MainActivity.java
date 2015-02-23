package com.jduplessis.results.activities;

import android.accounts.AccountManager;
import android.accounts.AccountManagerCallback;
import android.accounts.AccountManagerFuture;
import android.app.Fragment;
import android.app.FragmentManager;
import android.os.Bundle;
import android.support.v4.widget.DrawerLayout;
import android.support.v7.app.ActionBar;
import android.support.v7.app.ActionBarActivity;
import android.support.v7.widget.Toolbar;
import android.util.Log;
import android.view.Menu;
import android.view.MenuItem;

import com.jduplessis.results.R;
import com.jduplessis.results.api.Client;


public class MainActivity extends ActionBarActivity implements NavigationDrawerFragment.NavigationDrawerCallbacks{

    /**
     * Fragment managing the behaviors, interactions and presentation of the navigation drawer.
     */
    private NavigationDrawerFragment mNavigationDrawerFragment;

    /**
     * Used to store the last screen title. For use in {@link #restoreActionBar()}.
     */
    private CharSequence mTitle;
    private AccountManager mAccountManager;
    private Client mClient;

    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        mClient = Client.getInstance();

        setContentView(R.layout.activity_main);

        mNavigationDrawerFragment = (NavigationDrawerFragment)
                getSupportFragmentManager().findFragmentById(R.id.navigation_drawer);
        mTitle = getTitle();

        // Set up the drawer.
        mNavigationDrawerFragment.setUp(
                R.id.navigation_drawer,
                (DrawerLayout) findViewById(R.id.drawer_layout));

        mAccountManager = AccountManager.get(getBaseContext());
        mAccountManager.getAuthTokenByFeatures(LoginActivity.KEY_ACCOUNT_TYPE, LoginActivity.KEY_AUTH_TYPE, null, this, null, null, new AccountManagerCallback<Bundle>() {
            @Override
            public void run(AccountManagerFuture<Bundle> future) {
                Bundle bnd = null;
                try {
                    bnd = future.getResult();
                    final String authToken = bnd.getString(AccountManager.KEY_AUTHTOKEN);
                    mClient.setAuthToken(authToken);
                    Log.d("rc", "Set token");
                } catch (Exception e) {
                    e.printStackTrace();
                }
                onAuthReady();
            }
        }, null);
    }

    private void onAuthReady() {
        onNavigationDrawerItemSelected(0);
    }

    @Override
    public void onNavigationDrawerItemSelected(int position) {
        if(mClient.getAuthToken() == null) {
            return;
        }
        Fragment curFragment = null;
        switch (position) {
            case 0:
                // Results
                curFragment = new ResultsFragment();
                break;
            case 1:
                // Setup
                curFragment = new SetupFragment();
                break;
            case 2:
                // Log out
                mAccountManager.invalidateAuthToken(AccountManager.KEY_AUTHTOKEN, mClient.getAuthToken());
                recreate();
                return;

        }

        FragmentManager fragmentManager = getFragmentManager();
        fragmentManager.beginTransaction()
                .replace(R.id.container, curFragment)
                .commit();

        onSectionAttached(position);
    }

    public void onSectionAttached(int number) {
        switch (number) {
            case 1:
                mTitle = getString(R.string.title_section1);
                break;
            case 2:
                mTitle = getString(R.string.title_section2);
                break;
            case 3:
                mTitle = getString(R.string.title_section3);
                break;
        }
        getSupportActionBar().setTitle(mTitle);
    }

    public void restoreActionBar() {
        ActionBar actionBar = getSupportActionBar();
        actionBar.setDisplayShowTitleEnabled(true);
        actionBar.setTitle(mTitle);
    }
}
