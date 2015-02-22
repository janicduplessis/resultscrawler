//
//  HomeViewController.swift
//  
//
//  Created by Charles Vinette on 2015-02-13.
//
//

import UIKit

class HomeViewController: UIViewController {
    
    let client = Client.sharedInstance
    
    override func viewDidLoad() {
        super.viewDidLoad()
        // Do any additional setup after loading the view, typically from a nib.
        
        client.results("20151", callback: { (results) in
            if let results = results {
                
            }
        })
        
    }
    
    override func didReceiveMemoryWarning() {
        super.didReceiveMemoryWarning()
        // Dispose of any resources that can be recreated.
    }
    
    
}

