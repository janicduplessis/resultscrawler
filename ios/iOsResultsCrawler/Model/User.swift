//
//  User.swift
//  iOsResultsCrawler
//
//  Created by Janic Duplessis on 2015-02-21.
//  Copyright (c) 2015 App and Flow. All rights reserved.
//

import Foundation

struct User {
    var email: String?
    var firstName: String?
    var lastName: String?
}

struct LoginRequest {
    var email: String?
    var password: String?
    let deviceType = 1
}

struct LoginResponse {
    var status: Int?
    var token: String?
    var user: User?
}
