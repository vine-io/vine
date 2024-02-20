// MIT License
//
// Copyright (c) 2020 The vine Authors
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package template

var (
	HTMLWEB = `<!DOCTYPE html>
<html>
    <head>
        <title>{{title .Alias}} {{title .Type}}</title>
        <meta name="viewport" content="width=device-width, initial-scale=1">
        <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap.min.css" integrity="sha384-1q8mTJOASx8j1Au+a5WDVnPi2lkFfwwEAa8hDDdjZlpLegxhjVME1fgjWPGmkzs7" crossorigin="anonymous">
        <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap-theme.min.css" integrity="sha384-fLW2N01lMqjakBkx3l/M9EahuwpSfeNvV63J5ezn3uZzapT0u7EYsXMjQV+0En5r" crossorigin="anonymous">
    </head>
    <body>
        <nav class="navbar navbar-default">
            <div class="container">
                <div class="navbar-header">
                    <a class="navbar-brand" href="#">
                        {{title .Alias}} {{title .Type}}
                    </a>
                </div>
            </div>
        </nav>
        <div class="container">
            <div class="row">
                <div class="col-md-8">
                    <h1>{{title .Alias}} {{title .Type}}</h1>
                    <form class="{{.Name}}">
                        <div class="form-group">
                            <label>Enter your name</label>
                            <input type=text class="form-control" id="name" name="name" placeholder="John">
                         </div>
                        <button class="btn btn-default">Submit</button>
                    </form>
                </div>
            </div>
            <div class="row">
                <div class="col-md-8">
                    <div class="message"></div>
                </div>
            </div>
        </div>


        <script src="https://cdnjs.cloudflare.com/ajax/libs/jquery/2.2.2/jquery.min.js"></script>
        <script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/js/bootstrap.min.js" integrity="sha384-0mSbJDEHialfmuBBQP6A4Qrprq5OVfW37PRR3j5ELqxss1yVqOtnepnHVP9aJ7xS" crossorigin="anonymous"></script>

        <!-- You may want to store this in a separate file -->
        <script type="text/javascript">
            $(".{{.Name}}").submit(function(e) {
                e.preventDefault();

		var url = window.location.href.replace(/\/$/, "") + "/{{.Name}}/call";
                var data = $(this).serializeArray()[0];
                var name = data.value;
                if (name.length == 0) {
                    name = "John";
                };

                $.ajax({
                    method: "POST",
                    dataType: "json",
                    contentType: "application/json",
                    url: url,
                    data: JSON.stringify({name: name}),
                    success: function(data) {
                        $('.message').html('<h3>'+data.msg+'</h3>');
                    },
                });
            });
        </script>
    </body>
</html>
`
)
