this.navigator = {
	"userAgent": "1234"
};
this.document = false;
this.$jPlayer = function(fn) {
	var obj = {
		"ready": function(fn) { fn(); }
	}
	if (fn == "getInstance") {
		obj = function() {};
	}
	return obj;
};
this.jQuery = function() {
	return {
		"empty": function() {}
	}
};
var GlobalModel = ErrorController = ViewController = PlayerModel = Cookie = QueryParser ={ resetSingleton: function() {}};
var IVWController = {
	getInstance: function() {return "getInstance";}
}
var MediaCollection = function(vid, somebool, qual) {
	// console.log("video:" + vid);
	// console.log("someBool:" + somebool);
	// console.log("quality:" + qual);
	return {
		"addMediaStream": function(i1, i2, appUrl, path, def) {
			if(i1 == 0 && i2 == 2) {
				_rtmpUrl_found(appUrl, path)
			}
			// console.log("appUrl:"+appUrl);
			// console.log("path:"+path);
		},
		setSortierung: function() {},
		setPreviewImage: function() {},
		addMedia: function() {},

	}
}
var PlayerConfiguration = function() {
	return {
		setRepresentation: function() {},
		setBaseUrl: function() {},
		setSkinPathAndFileName: function() {},
		setAutoPlay: function() {},
		setShowOptions: function() {},
		setShowToolbar: function() {},
		setShowToolbarDownloadButtons: function() {},
		setShowToobarQualButtons: function() {},
		setShowToobarQualNavButtons: function() {},
		setZoomEnabled: function() {},
		setShowSettings: function() {},
		setModalTarget: function() {},
		setAddons: function() {},
		setNoSubtitelAtStart: function() {},
		setConvivaEnabled: function() {}
	}
}
var Player = function() {
	return false;
}