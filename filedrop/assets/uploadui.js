document.addEventListener("DOMContentLoaded", function(event) { 
    var _submit = document.getElementById('_submit'), 
    _file = document.getElementById('_file');
    
    var $pb = $('.progress .progress-bar');

    var upload = function(){

        if(_file.files.length === 0){
            return;
        }

        var data = new FormData();
        data.append('fupload', _file.files[0]);

        var request = new XMLHttpRequest();
        request.onreadystatechange = function(){
            if(request.readyState == 4){
                try {
                    var resp = JSON.parse(request.response);
                } catch (e){
                    var resp = {
                        status: 'error',
                        data: 'Unknown error occurred: [' + request.responseText + ']'
                    };
                }
                alert(resp.status + ': ' + resp.data);
            }
        };

        request.upload.addEventListener('progress', function(e){
            $pb.attr('data-transitiongoal', Math.ceil(e.loaded* 100/e.total)).progressbar({display_text: 'fill'});
        }, false);

        request.open('POST', 'upload');
        request.send(data);
    }

    _submit.addEventListener('click', upload);
});