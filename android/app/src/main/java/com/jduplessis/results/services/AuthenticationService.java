package com.jduplessis.results.services;

import android.app.Service;
import android.content.Intent;
import android.os.IBinder;

import com.jduplessis.results.authenticator.AccountAuthenticator;

/**
 * Created by Janic on 2015-01-18.
 */
public class AuthenticationService extends Service {
    @Override
    public IBinder onBind(Intent intent) {
        return new AccountAuthenticator(this).getIBinder();
    }
}
