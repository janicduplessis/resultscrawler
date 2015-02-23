//
//  Results.swift
//  iOsResultsCrawler
//
//  Created by Janic Duplessis on 2015-02-22.
//  Copyright (c) 2015 App and Flow. All rights reserved.
//

import Foundation

struct Results {
    var lastUpdate: String
    var classes: [Class]
}

struct Class {
    var id: String
    var name: String
    var group: String
    var results: [Result]
    var total: ResultInfo
    var finalGrade: String
}

struct Result {
    var name: String
    var normal: ResultInfo
    var weighted: ResultInfo
}

struct ResultInfo {
    var result: String
    var average: String
    var standardDev: String
}