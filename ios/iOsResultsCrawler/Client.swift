//
//  Client.swift
//  iOsResultsCrawler
//
//  Created by Charles Vinette on 2015-02-16.
//  Copyright (c) 2015 App and Flow. All rights reserved.
//

import Foundation

class Client {
    
    let baseURL :NSURL!
    let loginURL : NSURL!
    
    init(){
    
    baseURL = NSURL(string: "http://results.jdupserver.com/api/v1")
    
    loginURL = NSURL(string: "auth/login", relativeToURL:baseURL)
        
    }
    
    func login(email:String, passwordServer:String) {
        
        
        
        
        
    }
    






}
