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
        private static let baseURL = "https://results.jdupserver.com/api/v1"
        
        // stores the authentication token.
        static var authToken: String?
        
        // Login request.
        case Login(String, String)
        // Get results.
        case Results(String)
        // Register request.
        case Register(String,String,String,String)
        
        // URLRequestConvertible protocol.
        var URLRequest: NSURLRequest {
            // Returns the path, http method and parameters for the request.
            let (path: String, method: Alamofire.Method,  parameters: [String: AnyObject]?) = {
                switch self {
                case .Login (let email, let password):
                    let params: [String: AnyObject] = [
                        "email": email,
                        "password": password,
                        "deviceType": 1
                    ]
                    return ("/auth/login", .POST, params)
                case .Results(let session):
                    return ("/results/" + session, .GET, nil)
                    
                case .Register (let email, let password, let firstName,let lastName):
                    let params: [String:AnyObject] = [
                        "email":email,
                        "password":password,
                        "firstName":firstName,
                        "lastName":lastName,
                        "deviceType":1]
                    return ("/auth/register", .POST, params)
                }
            }()
            
            // Setup the URLRequest.
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
    
    // Singleton
    class var sharedInstance: Client {
        return _instance
    }
    private init() {}
    
    // If the client is authentified. Will be true if the login method was called successfully
    var authentified: Bool {
        return Router.authToken != nil
    }
    
    // Login logs in the user with his email and password.
    func login(email:String, password:String, callback:(LoginResponse?) -> Void) {
        Alamofire.request(Router.Login(email, password)).responseJSON { (_, _, data, error) in
            if(error != nil) {
                callback(nil)
                return
            }
            
            var json = JSON(data!)
            
            let status = json["status"].intValue
            let token = json["token"].stringValue
            let user = json["user"]
            let email = user["email"].stringValue
            let firstName = user["firstName"].stringValue
            let lastName = user["lastName"].stringValue
            
            Router.authToken = token
            
            callback(LoginResponse(
                status: LoginStatus(rawValue: status)!,
                token: token,
                user: User(email: email, firstName: firstName,lastName: lastName)
            ))
        }
    }
    
    func register (email:String, password:String, firstName:String, lastName:String, callback:(RegisterResponse?)->Void){
        Alamofire.request(Router.Register(email, password, firstName, lastName)).responseJSON {(_,_,data,error) in
            if(error != nil){
                callback(nil)
                return
            }
            var json = JSON(data!)
            
            let status = json ["status"].intValue
            let token = json ["token"].stringValue
            let user = json ["user"]
            let email = json ["email"].stringValue
            let firstName = user ["firstName"].stringValue
            let lastName = user ["lastName"].stringValue
            
            Router.authToken = token
            
            callback(RegisterResponse(status: RegisterStatus(rawValue: status)!, token: token, user: User(email: email, firstName: firstName, lastName: lastName)))
        
        }
    }
    
    func results(session:String, callback:(Results?) -> Void) {
        
        func parseResultInfo(json: JSON) -> ResultInfo {
            let result = json["result"].stringValue
            let average = json["average"].stringValue
            let stdDev = json["standardDev"].stringValue
            return ResultInfo(result: result, average: average, standardDev: stdDev)
        }
        
        func parseResults(json: JSON) -> [Result] {
            var results = [Result]()
            for (index: String, subJson: JSON) in json {
                let name = json["name"].stringValue
                let normal = parseResultInfo(json["normal"])
                let weighted = parseResultInfo(json["weighted"])
                results.append(Result(name: name, normal: normal, weighted: weighted))
            }
            return results
        }
        
        Alamofire.request(Router.Results(session)).responseJSON { (_, _, data, error) -> Void in
            if(error != nil) {
                callback(nil)
                return
            }
            
            var json = JSON(data!)
            let lastUpdate = json["lastUpdate"].stringValue
            var classes = [Class]()
            for (index: String, subJson: JSON) in json["classes"] {
                let id = subJson["id"].stringValue
                let name = subJson["name"].stringValue
                let group = subJson["group"].stringValue
                let finalGrade = subJson["final"].stringValue
                classes.append(Class(id: id, name: name, group: group, results: parseResults(json["results"]), total: parseResultInfo(subJson), finalGrade: finalGrade))
            }
            
            callback(Results(lastUpdate: lastUpdate, classes: classes))
        }
    }
}
