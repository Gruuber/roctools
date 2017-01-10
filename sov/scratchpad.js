// ==UserScript==
// @name        Sell Of Value
// @namespace   splopi
// @description Getts sell value from recons
// @include     https://ruinsofchaos.com/inteldetail.php*
// @include     https://ruinsofchaos.com/stats.php*
// @version     1
// @grant       none
// ==/UserScript==

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
  
  console.log("Intel scan completed");
}

//Call this function if in the stats page
/**
 * This will function will add the sell off value on the user page
 * If no value is found, a text will appear asking to recon
**/
function stats (){
  console.log("Stats report completed");
}

//This is a self executing anonymous function, 
//Used because it is easy to stop its execution whenever just by using return
(function(){
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
