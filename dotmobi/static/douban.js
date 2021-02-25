void((function(){
    var url = window.location.href;
    var text = $('#wrapper h1 span').text();
    if (text) {
        window.location.href = 'http://dotmobi.sinaapp.com/search/' + text;
        return false;
    }
})())
