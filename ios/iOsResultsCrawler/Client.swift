//
//  Client.swift
//  iOsResultsCrawler
//
//  Created by Charles Vinette on 2015-02-16.
//  Copyright (c) 2015 App and Flow. All rights reserved.
//

import Foundation
import Alamofire
import SwiftyJSON

class Client {
    
    private enum Router: URLRequestConvertible {
        private static let baseURL = "http://results.jdupserver.com/api/v1"
        
        static var authToken: String?
        
        case Login(String, String)
        
        var URLRequest: NSURLRequest {
            let (path: String, method: Alamofire.Method,  parameters: [String: AnyObject]) = {
                switch self {
                case .Login (let email, let password):
                    let params: [String: AnyObject] = [
                        "email": email,
                        "password": password,
                        "deviceType": 1
                    ]
                    return ("/auth/login", .POST, params)
                }
            }()
            
            let url = NSURL(string: Router.baseURL)
            let urlRequest = NSMutableURLRequest(URL: url!.URLByAppendingPathComponent(path))
            urlRequest.HTTPMethod = method.rawValue
            if let authToken = Router.authToken {
                urlRequest.addValue(authToken, forHTTPHeaderField: "X-Access-Token")
            }
            let encoding = Alamofire.ParameterEncoding.JSON
            
            return encoding.encode(urlRequest, parameters: parameters).0
        }
    }
    
    var authentified: Bool {
        return Router.authToken != nil
    }
    
    func login(email:String, password:String, callback:(LoginResponse?) -> Void) {
        Alamofire.request(Router.Login(email, password)).responseJSON { (_, _, json, error) in
            if(error != nil) {
                callback(nil)
                return
            }
            
            
            
            
            println(json)
        }
    }
}
