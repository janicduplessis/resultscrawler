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
    
    @IBOutlet weak var loading: UIActivityIndicatorView!
    
    @IBAction func backLoginPage(sender: UIButton) {
        
        let loginViewController = LoginViewController()
        self.dismissViewControllerAnimated(true , completion: nil)  }
    
    

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
        
        self.loading.startAnimating()
        
        if(firstName != "" && lastName != "" && email != "" && password != ""){
            
            client.register(email, password: password, firstName: firstName, lastName: lastName, callback: {(response) in
                
                if let response = response{
                    if response.status == RegisterStatus.Ok{
                       self.loading.stopAnimating()
                        
                        let signupSuccess = UIAlertController(title: "Compte crée!", message: nil, preferredStyle: .Alert)
                        
                        let retour = UIAlertAction(title: "Retour", style: .Default, handler: { (retour) -> Void in
                        self.dismissViewControllerAnimated(true , completion: nil)
                            
                            
                        })
                        
                        signupSuccess.addAction(retour)
                        self.presentViewController(signupSuccess, animated: true, completion: nil)
                        }
                }else{
                    self.loading.stopAnimating()
                    let signupFail = UIAlertController(title: "Erreur", message: "Champs vides", preferredStyle: .Alert)
                    let retourSignup = UIAlertAction(title: "Ok", style: .Default, handler: { (retourSignup) -> Void in
                        self.dismissViewControllerAnimated(true , completion: nil)
                    })
                    signupFail.addAction(retourSignup)
                    self.presentViewController(signupFail, animated: true, completion: nil)
                    
                }
            })
        }
        else{
            self.loading.stopAnimating()
            let signupFail = UIAlertController(title: "Erreur", message: "Champs vides", preferredStyle: .Alert)
            let retourSignup = UIAlertAction(title: "Ok", style: .Default, handler: { (retourSignup) -> Void in
                self.dismissViewControllerAnimated(true , completion: nil)
            })
            signupFail.addAction(retourSignup)
            self.presentViewController(signupFail, animated: true, completion: nil)
            
        }
    }

}
