//
//  Moodle.swift
//  iOsResultsCrawler
//
//  Created by Charles Vinette on 2015-03-04.
//  Copyright (c) 2015 App and Flow. All rights reserved.
//

import Foundation
import Alamofire
import SwiftyJSON

class Moodle{
    
    
    func loginMoodle(codeMS:String, motDePasse: String)->Void{
        
        
        
        let username = codeMS //TEXTFIELD A VENIR
        let password = motDePasse //TEXTFIELD A VENIR
        
        
        Alamofire.request(.POST," https://www.moodle.uqam.ca/login/token.php?username=\(username)&password=\(password)&service=moodle_mobile_app").responseJSON { (_, _ , data, error) in
            if(error != nil){
                //error
            }
            
            var json = JSON(data!)
            
            let token = json["token"].stringValue
        }
        
    }
    
    
    
    
    
    
    
    
    
    
    
}
