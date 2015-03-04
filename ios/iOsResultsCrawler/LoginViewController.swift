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
    
    
    
    //ERREUR AU NIVEAU DE LA CONNECTION, NE S'ARRETE PAS MEME SI COMBINAISON EST MAUVAISE!!!!
    @IBAction func connect() {
        let code = codeTextField.text
        let nip = nipTextField.text
        self.loadingLogin.startAnimating()
        
        if code != "" && nip != "" {
            client.login(code, password: nip, callback: { (response) in
                if let response = response {
                    if response.status == LoginStatus.Ok {
                        // Good login.
                        self.loadingLogin.stopAnimating()
                        let homeViewController = self.storyboard!.instantiateViewControllerWithIdentifier("HomeViewController") as HomeViewController
                        
                        self.showViewController(homeViewController, sender: self)
                    }
                } else {
                    self.loadingLogin.stopAnimating()
                    let badLogin = UIAlertController(title: "Échec de connexion", message: "La combinaison du email et du mot de passe n'est pas bonne", preferredStyle: .Alert)
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
