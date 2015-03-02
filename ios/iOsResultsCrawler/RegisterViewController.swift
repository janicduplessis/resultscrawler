//
//  RegisterViewController.swift
//  iOsResultsCrawler
//
//  Created by Charles Vinette on 2015-03-02.
//  Copyright (c) 2015 App and Flow. All rights reserved.
//

import UIKit

class RegisterViewController: UIViewController {
    
    @IBOutlet weak var firstNameTextField: UITextField!

    @IBOutlet weak var lastNameTextField: UITextField!
    
    @IBOutlet weak var emailTextField: UITextField!
    
    @IBOutlet weak var passwordTextField: UITextField!
    
    
    let client = Client.sharedInstance
    
    
    
    override func viewDidLoad() {
        
        super.viewDidLoad()
        
    
    
    }
    
    
    
    override func didReceiveMemoryWarning() {
        super.didReceiveMemoryWarning()
    
    
    }
    
    @IBAction func register(sender: UIButton) {
        let firstName = firstNameTextField.text
        let lastName = lastNameTextField.text
        let email = emailTextField.text
        let password = passwordTextField.text
        
        if(firstName != nil && lastName != nil && email != nil && password != nil){
            
            client.register(email, password: password, firstName: firstName, lastName: lastName, callback: {(response) in
                if let response = response{
                    if response.status == RegisterStatus.Ok{
                        let loginViewController = self.storyboard!.instantiateViewControllerWithIdentifier("LoginViewController") as LoginViewController
                        
                        self.showViewController(loginViewController, sender: self)
                    }
                }else{
                    //erreur
                }
            })
        }
    }

}
