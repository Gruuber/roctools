// ==UserScript==
// @name        Sell Of Value
// @namespace   splopi
// @description Getts sell value from recons
// @include     https://ruinsofchaos.com/inteldetail.php*
// @include     https://ruinsofchaos.com/stats.php*
// @version     2.02
// @grant       none
// ==/UserScript==
//This is a self executing anonymous function, 
//Used because it is easy to stop its execution whenever just by using return
(function () {
  
  //Global variable
  var nameToPriceMap = {
    'Dagger': 800,
    'Maul': 12000,
    'Blade': 160000,
    'Excalibur': 800000,
    'Sai': 800,
    'Shield': 12000,
    'Mithril': 160000,
    'Dragonskin': 800000,
    'Cloak': 40000,
    'Hook': 80000,
    'Pickaxe': 240000,
    'Horn': 40000,
    'Guard Dog': 80000,
    'Torch': 240000
  };
  
  //We can also guess the price from the strength. Dagger Sai Cloak and Horn share the same strength, so we would need to do some more testing before guessing the price
  var strToPriceMap = {
    '300': 12000,
    '3,000': 16000,
    '12,000': 800000,
    '50': 80000,
    '120': 240000
  };
  
  //Call this function if the intel page
  /**
   * This function will gather inforamtion about the user's weapons and stores the SOV in a cookie
  **/
  function inteldetail() {
    /*
    * This grabs the first part of the page that inculdes information about the report
    ********************************************************************************
    * EX: Your spy sneaks into someone's base to attempt to gather intelligence.   *
    * Your spy enters undetected. The spy gathers the following information:       *
    ********************************************************************************
    * Using this information we will determin if the intel report is a successful recon, 
    * or something else we don't care about
    */
    //Local variables
    //This will store all the weapons for that user. Will also be used to store in a cookie for use outside of the intel page
    var weaponsArray = [
    ];
    //The current estimated sell value without calculating the weapons that cannot be  indentified
    var cesov = 0;
    //This grabs the acutal object and not the text
    //This might be bad cause it depends on the class. Will figure out a way the description using a different method
    var descriptionObject = $('.td.c');
    //Grab the text from the object 
    var descriptionText = descriptionObject.html();
    /**
     * Sample out put of the descriptionText
     * console.log(descriptionText);

      Your 10 spies sneak into <a href="stats.php?id=28380">someone</a>'s base to attempt to gather intelligence.<br>
    <br>
    Your spies enter undetected.
    <font size="3" color="RED">2</font> of your spies were a little slow and had to become an hero in order to evade capture and interrogation.
    The spies gather the following information:

     * We will be using the use ID to reference the player
    **/
    //Search for "gathers the following information" to determin if this is the info we need
    if (descriptionText.indexOf('gathers the following information') == 0) {
      //This is not a valid log, and we need to stop the script, 
      //will comment out the log to keep things clean
      //console.log("This is not a valid intel log");
      //Return will cause the function to finish executing
      return;
    }    //Use regular expression to match and extract the userID

    var userIDregex = /stats\.php\?id=([0-9]+)/;
    var match = descriptionText.match(userIDregex);
    if (match.length < 2) {
      //Either code has changed, or I did something wrong, end the scrip for now
      return;
    }    //We store the user id, it will be later used to store the remaining values
    var userID = match[1];

    //Now we grab the weapons type
    //We start by grabbig all the tables from the page
    var tables = $('table.sep');
    //Iterate over all the retrieved table to find the one with the header title "Weapons"
    var weaponsObj = false;
    tables.each(function () {
      //make a jQuery object
      var tableObj = $(this);
      var tbodyObj = tableObj.children();
      //Check if the first element is a table body
      //I do not think this is needed. And that check can be bypassed
      if (tbodyObj[0].localName == 'tbody') {
        //We now fetch the first row from the table
        var trObj = tbodyObj.children().first();
        //Each row is formed of multiple cells, in the case of the header
        //It only has one cell, that contains the title
        //We fetch the first cell and check the title name
        var headerCell = trObj.children().first();
        //We check the value inside to figure out if this is the table we need
        //3 equals also matches the type
        if (headerCell.html() === 'Weapons') {
          //Once we found the Weapons table, we store the object and return the inner function
          //This return will not end the entire script, it will only end the function we are currently in
          weaponsObj = tableObj;
          return;
        }
      }
    });
    //close the script if we did not find the Weapons table
    if (weaponsObj === false) {
      return;
    }    //Get rows from the table.
    //The structure is as follow, Table > tablebody > rows > cells. Also skip the first 2 elemnts since they are headers.
    //gt is zero based, gt(0) will skip the first, gt(1) will skip the first 2

    var weaponRows = weaponsObj.children().children(':gt(1)');
    //Get the first row for the title, we will use that later to output our sell off value
    var titleRow = weaponsObj.children().children(':nth-child(1)');
    //Get the titleCell
    titleCell = $(titleRow.children() [0]);
    //Looping over all the rows, each row is a description of a weapon held by user
    weaponRows.each(function () {
      //Create object from row
      var weaponRow = $(this);
      //Get the cells from the row, there are 4 cells, name, type, quantity and strength
      var weaponCells = weaponRow.children();
      //Create an object of the cell, then extract the value from each cell.
      var name = $(weaponCells[0]).html();
      //We need to strip the link from the name that links to a sab. Regular expressions are your friend
      //Sample name text without the link stripped
      //<a href="attack.php?sab=WhoCares&amp;weapon=Maul">Maul</a>
      //Keep in mind that the value can be "???" we need to account for that. I used ternary  operator here
      name = (name === '???') ? '???' : name.match(/>(.*)<\/a>/) [1];
      var type = $(weaponCells[1]).html();
      //Remove the comma from the quantity and transform the string to an integer
      var quantity = ($(weaponCells[2]).html() === '???') ? '???' : parseInt($(weaponCells[2]).html().replaceAll(',', ''));
      var curStrength = $(weaponCells[3]).html();
      //Splitting the curStrength to get the current damage and the max strength
      //The max strength will help us determin the sell value of the item even if we are lacking information
      //Check for ??? as well.
      //We add place holder values
      var damage = '???';
      var strength = '???';
      if (curStrength !== '???') {
        var strengthTok = curStrength.split('/');
        damage = strengthTok[0];
        strength = strengthTok[1];
      }      //Will now create an object for this weapon, this object will be used to store it in a cookie for later use.

      var wep = {
        'name': name,
        'type': type,
        'quantity': quantity,
        'damage': damage,
        'strength': strength
      };
      //Add that object to the weapon array
      weaponsArray.push(wep);

    });
    //Try to get the user from the object stored in the cookie, in the case the cookie does not exist, create a new users object
    //If the user does not exist, create a user object.
    var UsersObj = {};
    //Make the string back to a JSON object
    var splopiCookie = JSON.parse(readCookie("usersWeapons"));
    
    if(splopiCookie !== null){
      UsersObj = splopiCookie;
    }
    
    //If the user exsists, we will attempt to combine the old data with the new data to get the full SOV
    if(typeof UsersObj[userID] !== "undefined"){
      var preWeapons = UsersObj[userID].weapons;
      //Now for some ungly programing using double for loop to loop over everything.
      //Will first loop over the weaponsArray, and for each item. If the value cannot be guessed. We try to match data from the old list from older recons
      for (var i = 0; i < weaponsArray.length ; i++) {
        //This if checks if the value can be calulated
        //Get the weapon object. Also minimizes the large the var name
        var wo = weaponsArray[i];
        if( wo.quantity === '???' || !( wo.name !== '???' || (wo.strength !== '???' && wo.strength !== '30') || (wo.strength === '30' && wo.type !== '???'))){
          //Condition 1 : First we check if the quantity is available. 
          //Condition 2 : Then we check if all of the other criterea used below in getSellValueOf() are false. 
          //If either of condition 1 or 2 are false, we search for extra information
          //Loop over the old data and try to find a match
          for (var x = 0; c < preWeapons.length ; x++) {
            //pwo stands for previous weapons object
            var pwo = preWeapons[i];
            //This variable will be used as a flag to signal that we found a match
            var matchFlag = false;
            //First we try to match using the name
            if(wo.name !== "???" && pwo.name !== "???" && wo === pwo.name){
              //The above statement translate to
              //name of current object is NOT ??? and name of pre object is not ??? and the current object and previous object have the name name
              matchFlag = true;
            }
            //The second way we match the item for sure if the compare the strength and the type
            if(wo.strength !== "???" && wo.type !== "???" && pwo.strength !== "???" && pwo.type !== "???" && wo.strength === pwo.strength && wo.type === pwo.type){
              //The above is too long to translate to english
              matchFlag = true;
            }
            
            //We can also implement one more technique where we comapre the types of elements found to the types of elemts not found to fill in the rest
            //But I am too lazy and you should just recon more that once and increase you SPY stats
            
            //If we found a match, we replace the missing values
            if(matchFlag){
              if(wo.name === "???" && pwo.name !== "???"){
                wo.name = pwo.name;
              }
              if(wo.type === "???" && pwo.type !== "???"){
                wo.type = pwo.type;
              }
              if(wo.quantity === "???" && pwo.quantity !== "???"){
                wo.quantity = pwo.quantity;
              }
              if(wo.name === "???" && pwo.name !== "???"){
                wo.strength = pwo.strength;
              }
            }
            //Since we found our match, we can break out of the loop
            break;
            
          }
        }
      }
    }
    
    //Loop over the array, and caluculate the weapons sell off value
    for (var c = 0; c < weaponsArray.length ; c++) {
      cesov += getSellValueOf(weaponsArray[c]);
    }
          
    //Create a new user object with a new array 
    var UserObj = {"weapons" : weaponsArray , "SOV" : cesov};
    UsersObj[userID] = UserObj;
    //Delete pre cookies before writing a new one
    eraseCookie("splopiCookie");
    //Store the new data in the cookie. Also make a string out of the object to be able to store it
    createCookie("splopiCookie" , JSON.stringify(UsersObj) , 1);
    
    titleCell.html(titleCell.html() + '. Estimated SOV ' + cesov.toLocaleString());
  }  
  
  //Call this function if in the stats page
  /**
   * This will function will add the sell off value on the user page
   * If no value is found, a text will appear asking to recon
  **/

  function stats() {
    //We check if our cookie is set.
    var splopiCookie = JSON.parse(readCookie("splopiCookie"));
    if(splopiCookie !== null){
      //If so, we fetch the userId from the url. We need all the didgits after the stats.php?id= part
      var userIdTok = pageURL.match(/stats\.php\?id=([0-9]*)/);
      if(userIdTok.length === 2){
        var userId = userIdTok[1];
        //Check if we have information about that user
        if(typeof splopiCookie[userId] !== "undefined"){
          //If so, we need to read the sell off value, and add it to the user page
          var sov = splopiCookie[userId].SOV;
          //Now we need to fetch the table body that includes the name
          //Please let this be and easy fetch!
          var tables = $(".sep.f");
          //This returned all the table with class .sep and .f
          //Btw, if the admins decide to change the CSS, this would need to be rewritten!!
          //There has to be a better way!!!
          var goldTable = tables[1];
          //Create an object from the table
          var gtObj = $(goldTable);
          //Create the row that needs to be added to the table. I used single quotes to avoid breaking the double quotes
          var sovRow = $('<td class="lg" style="width: 25%"><b>SOV:</b></td><td class="lg" style="width: 75%">'+sov.toLocaleString()+'</td>');
          //And append the sov row to the tbody
          gtObj.children().append(sovRow);
        }
      }else{
        console.log("STOP TRYING TO BREAK MY SCRIPT");
      }
    }
  }  
  
  //This function calculates the sell value of weapons passed. Will return a value of zero in the case the weapon cannot be indentidfied
  function getSellValueOf(wepObj) {
    //We need to iterate 
    var value = 0;
    if (wepObj.quantity !== '???') {
      //Only attempt to calculate the value if we know the auantity
      if (wepObj.name !== '???') {
        value = nameToPriceMap[wepObj.name] * wepObj.quantity;
      } else if (wepObj.strength !== '???' && wepObj.strength !== '30') {
        //If we do not know the name, we can still calcualte the value using the strength and type combination
        value = strToPriceMap[wepObj.strength] * wepObj.quantity;
      } else if (wepObj.strength === '30' && wepObj.type !== '???') {
        //In the case the strength 30, we would need to check the type to get the value
        value = (wepObj.type === 'Attack' || wepObj.type === 'Defense') ? 800 * wepObj.quantity : 40000 * wepObj.quantity;
      }
    }
    return value;
  }
  
  //Functions to read write and erase cookies
  //I just got these from the internet!! I love you internet
  function createCookie(name, value, days) {
    if (days) {
      var date = new Date();
      date.setTime(date.getTime() + (days * 24 * 60 * 60 * 1000));
      var expires = '; expires=' + date.toUTCString();
    } 
    else var expires = '';
    document.cookie = name + '=' + value + expires + '; path=/';
  }
  
  function readCookie(name) {
    var nameEQ = name + '=';
    var ca = document.cookie.split(';');
    for (var i = 0; i < ca.length; i++) {
      var c = ca[i];
      while (c.charAt(0) == ' ') c = c.substring(1, c.length);
      if (c.indexOf(nameEQ) == 0) return c.substring(nameEQ.length, c.length);
    }
    return null;
  }
  
  function eraseCookie(name) {
    createCookie(name, '', - 1);
  }  //Get page URL 
  
  
  String.prototype.replaceAll = function(search, replacement) {
    var target = this;
    return target.split(search).join(replacement);
  };
  
  //The following will just call the correct function based on the page name
  var pageURL = window.location.href;
  //Check which function to call.
  if (pageURL.match(/stats\.php/)) {
    stats();
  } else if (pageURL.match(/inteldetail\.php/)) {
    inteldetail();
  } else {
    console.log('Page not currently supported');
  }
  
})()