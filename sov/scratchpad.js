// ==UserScript==
// @name        Sell Of Value
// @namespace   splopi
// @description Getts sell value from recons
// @include     https://ruinsofchaos.com/inteldetail.php*
// @include     https://ruinsofchaos.com/stats.php*
// @version     1
// @grant       none
// ==/UserScript==

//This is a self executing anonymous function, 
//Used because it is easy to stop its execution whenever just by using return
(function(){

  //Global variable
  var nameToPriceMap = {"Dagger" : 800 , "Maul" : 12000 , "Blade" : 16000 , "Excalibur" : 800000 , "Sai" : 800 , "Shield" : 12000 , "Mithril" : 16000 , "Dragonskin" : 800000 , "Cloak" : 40000 , "Hook" : 80000 , "Pickaxe" : 240000 , "Horn" : 40000 , "Guard Dog" : 80000 , "Torch" : 240000  };
  
  
  //Call this function if the intel page
  /**
   * This function will gather inforamtion about the user's weapons and stores the SOV in a cookie
  **/
  function inteldetail(){

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
    var weaponsArray = [];
    //The current estimated sell value without calculating the weapons that cannot be  indentified
    var cesov = 0;

    //This grabs the acutal object and not the text
    //This might be bad cause it depends on the class. Will figure out a way the description using a different method
    var descriptionObject = $(".td.c");

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
    if( descriptionText.indexOf("gathers the following information") == 0){
      //This is not a valid log, and we need to stop the script, 
      //will comment out the log to keep things clean
      //console.log("This is not a valid intel log");
      //Return will cause the function to finish executing
      return;
    }

    //Use regular expression to match and extract the userID
    var userIDregex = /stats\.php\?id=([0-9]+)/;
    var match =  descriptionText.match(userIDregex);

    if(match.length < 2){
      //Either code has changed, or I did something wrong, end the scrip for now
      return;
    }
    //We store the user id, it will be later used to store the remaining values
    var userID = match[1];

    //Now we grab the weapons type
    //We start by grabbig all the tables from the page
    var tables = $("table.sep");
    //Iterate over all the retrieved table to find the one with the header title "Weapons"
    var weaponsObj = false ;

    tables.each(function(){
      //make a jQuery object
      var tableObj = $(this);
      var tbodyObj = tableObj.children();

      //Check if the first element is a table body
      //I do not think this is needed. And that check can be bypassed
      if(tbodyObj[0].localName == "tbody"){
        //We now fetch the first row from the table
        var trObj = tbodyObj.children().first();
        //Each row is formed of multiple cells, in the case of the header
        //It only has one cell, that contains the title
        //We fetch the first cell and check the title name
        var headerCell = trObj.children().first();

        //We check the value inside to figure out if this is the table we need
        //3 equals also matches the type
        if(headerCell.html() === "Weapons"){
          //Once we found the Weapons table, we store the object and return the inner function
          //This return will not end the entire script, it will only end the function we are currently in
          weaponsObj = tableObj;
          return;
        }
      }
    });

    //close the script if we did not find the Weapons table
    if(weaponsObj === false){
      return;
    }

    //Get rows from the table.
    //The structure is as follow, Table > tablebody > rows > cells. Also skip the first 2 elemnts since they are headers.
    //gt is zero based, gt(0) will skip the first, gt(1) will skip the first 2
    var weaponRows = weaponsObj.children().children(":gt(1)");

    //Looping over all the rows, each row is a description of a weapon held by user
    weaponRows.each(function(){
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
      name = (name === "???") ? "???" : name.match(/>(.*)<\/a>/)[1];
      var type = $(weaponCells[1]).html();
      var quantity = $(weaponCells[2]).html();
      var curStrength = $(weaponCells[3]).html();
      //Splitting the curStrength to get the current damage and the max strength
      //The max strength will help us determin the sell value of the item even if we are lacking information
      //Check for ??? as well.
      //We add place holder values
      var damage = "???";
      var strength = "???";

      if (curStrength !== "???"){
       var strengthTok = curStrength.split("/");
       damage = strengthTok[0];
       strength = strengthTok[1];
      }

      //Will now create an object for this weapon, this object will be used to store it in a cookie for later use.
      var wep = {"name" : name , "type" : type , "quantity" : quantity  , "damage" : damage , "strength" : strength };
      cesov += getSellValueOf(wep);

      //Add that object to the weapon array
      weaponsArray.push(wep);
    });
    console.log(weaponsArray);
  }

  //Call this function if in the stats page
  /**
   * This will function will add the sell off value on the user page
   * If no value is found, a text will appear asking to recon
  **/
  function stats (){
    console.log("Stats report completed");
  }

  //This function calculates the sell value of weapons passed. Will return a value of zero in the case the weapon cannot be indentidfied
  function getSellValueOf(wepObj){
    //We need to iterate 
    if(1){

    }

    return 0;
  }


  //Get page URL 
  var pageURL = window.location.href ;
 
  //Check which function to call.
  if(pageURL.match(/stats\.php/)){
    stats();
  }else if(pageURL.match(/inteldetail\.php/)){
    inteldetail();
  }else{
   console.log("Page not currently supported"); 
  }
  
})()
