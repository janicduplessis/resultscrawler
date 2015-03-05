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
    
    @IBOutlet weak var loadingLogin: UIActivityIndicatorView!
    
    let client = Client.sharedInstance
    
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
        let code_perm = codeTextField.text
        let nip = nipTextField.text
        self.loadingLogin.startAnimating()
        
        if code_perm != "" && nip != "" {
            client.login(code_perm, nip: nip, callback: { (response) in
                if let response = response{
                    
                        self.loadingLogin.stopAnimating()
                        let homeViewController = self.storyboard!.instantiateViewControllerWithIdentifier("HomeViewController") as HomeViewController
                    
                    
                        
                        self.showViewController(homeViewController, sender: self)
                    
                } else {
                    self.loadingLogin.stopAnimating()
                    let badLogin = UIAlertController(title: "Échec de connexion", message: "La combinaison du code permanent et du nip n'est pas bonne", preferredStyle: .Alert)
                    let reessayer = UIAlertAction(title: "Réessayer", style: .Default, handler: { (reessayer) -> Void in
                        self.dismissViewControllerAnimated(true , completion: nil)
                    })
                    badLogin.addAction(reessayer)
                    
                    self.presentViewController(badLogin, animated: true, completion: nil)
                    
                }
            })
        }
    }
    

}
