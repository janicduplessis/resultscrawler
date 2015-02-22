//
//  ViewController.swift
//  iOsResultsCrawler
//
//  Created by Charles Vinette on 2015-02-09.
//  Copyright (c) 2015 App and Flow. All rights reserved.
//

import UIKit

class LoginViewController: UIViewController {
    
    
    @IBOutlet weak var LoginScreenImage: UIImageView!

    @IBOutlet weak var codeTextField: UITextField!
    
    @IBOutlet weak var nipTextField: UITextField!
    
    var client = Client()
    
    override func viewDidLoad() {
        super.viewDidLoad()
        // Do any additional setup after loading the view, typically from a nib.
        
        LoginScreenImage.image = UIImage(named: "UQAMLOGO")
    }
    
    override func didReceiveMemoryWarning() {
        super.didReceiveMemoryWarning()
        // Dispose of any resources that can be recreated.
    }
    
    @IBAction func connect() {
        let code = codeTextField.text
        let nip = nipTextField.text
        
        if code != nil && nip != nil {
            client.login(code, password: nip, callback: { (response) in
                
            })
        }
    }
}
