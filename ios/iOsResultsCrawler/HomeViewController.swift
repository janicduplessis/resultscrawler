//
//  HomeViewController.swift
//  
//
//  Created by Charles Vinette on 2015-02-13.
//
//

import UIKit

class HomeViewController: UIViewController {
    

    @IBOutlet weak var schedule: UIImageView!
    
    @IBOutlet weak var courses: UIImageView!
    
    @IBOutlet weak var email: UIImageView!
    
    @IBOutlet weak var grades: UIImageView!
    
    
    let client = Client.sharedInstance
    
    override func viewDidLoad() {
        
        
        
        super.viewDidLoad()
        // Do any additional setup after loading the view, typically from a nib.
        
        schedule.image = UIImage(named:"schedule")
        courses.image = UIImage(named: "courses")
        email.image = UIImage(named:"mail")
        grades.image = UIImage(named:"grades")
        
        
        

        
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

