//
//  User.swift
//  iOsResultsCrawler
//
//  Created by Janic Duplessis on 2015-02-21.
//  Copyright (c) 2015 App and Flow. All rights reserved.
//

import Foundation

struct User {
    var email: String
    var firstName: String
    var lastName: String
}

enum LoginStatus: Int {
    case Ok
    case Invalid
}

struct LoginResponse {
    var status: LoginStatus
    var token: String
    var user: User
}
