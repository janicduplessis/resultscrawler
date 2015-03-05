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

private let _instance = Client()


class Client {
    // Router is used to do a request to the server.
    private enum Router: URLRequestConvertible {
        private static let baseURL = "https://mobile.uqam.ca/portail_etudiant"
        
        
        // stores the authentication token.
         static var code_perm: String?
         static var nip:String?
        
        // Login request.
        case Login(String, String)
       
        
        // URLRequestConvertible protocol.
        var URLRequest: NSURLRequest {
            // Returns the path, http method and parameters for the request.
            var (path: String, method: Alamofire.Method,  parameters: [String: AnyObject]) = {
                switch self {
                case .Login (let code_perm, let nip):
                    let params: [String: AnyObject] = [
                        "code_perm": code_perm,
                        "nip": nip,
                        ]
                    return ("/proxy_dossier_etud.php", .POST, params)
                
                    
                }
            }()
            
            // Setup the URLRequest.
            let url = NSURL(string: Router.baseURL)
            let urlRequest = NSMutableURLRequest(URL: url!.URLByAppendingPathComponent(path))
            urlRequest.HTTPMethod = method.rawValue
            if let code_perm = Router.code_perm {
                if let nip = Router.nip{
                    
                parameters["nip"] = nip
                parameters["code_perm"] = code_perm
                }
            }

            let encoding = Alamofire.ParameterEncoding.PropertyList(NSPropertyListFormat.XMLFormat_v1_0, 0)
            return encoding.encode(urlRequest, parameters: parameters).0
        }
    }
    
    // Singleton
    class var sharedInstance: Client {
        return _instance
    }
    private init() {}
    

    
    // Login logs in the user with his email and password.
    func login(code_perm:String, nip:String, callback:(LoginResponse?) -> Void) {
        Alamofire.request(Router.Login(code_perm, nip)).responseJSON { (_, _, data, error) in
            if(error != nil) {
                println(error)
                callback(nil)
                return
            }
            
            var json = JSON(data!)
            
            let prenom = json["socio"]["prenom"].stringValue
            let nom = json["socio"]["nom"].stringValue
            
            Router.code_perm = code_perm
            Router.nip = nip
           
            callback(LoginResponse(
                user: User(prenom: prenom, nom: nom)
                
            ))
        }
    }
}
