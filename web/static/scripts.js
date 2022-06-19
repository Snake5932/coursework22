let setHeaderBtn = function() {
    let div1 = document.createElement('div');
    div1.className = 'u-btn-2';
    let a1 = document.createElement('a')
    let div2 = document.createElement('div');
    div2.className = 'u-btn-2';
    let a2 = document.createElement('a')
    let head = document.getElementById('header')
    let xhr = new XMLHttpRequest();
    xhr.open('GET', '/api/checklogin', true);
    xhr.onreadystatechange = function() {
        if (this.status !== 200) {
            a1.innerText = 'Войти'
            a1.href = '/login'
            a1.className = 'u-form-group u-form-submit u-border-none u-btn u-btn-submit u-button-style u-text-hover-custom-color-3'
            div1.append(a1)
            head.append(div1)
            a2.innerText = 'Зарегистрироваться'
            a2.href = '/register'
            a2.className = 'u-form-group u-form-submit u-border-none u-btn u-btn-submit u-button-style u-text-hover-custom-color-3'
            div2.append(a2)
            head.append(div2)
        } else {
            let prefix = '';
            if (this.responseText === 'admin') {
                prefix = '/admin';
            }
            a2.innerText = 'Аккаунт'
            a2.href = prefix + '/account'
            a2.className = 'u-form-group u-form-submit u-border-none u-btn u-btn-submit u-button-style u-text-hover-custom-color-3'
            div2.append(a2)
            head.append(div2)
            a1.innerText = 'Выйти'
            a1.onclick = logout
            a1.className = 'u-form-group u-form-submit u-border-none u-btn u-btn-submit u-button-style u-text-hover-custom-color-3'
            div1.append(a1)
            head.append(div1)
        }
    };
    xhr.send();
}

let mailUnique = false
let nicknameUnique = false
let samePass = false
let page = 0
let book_page = 0
let user_book_page = 0
let gen_book_page = 0
let source_page = 0
let book_page_num = 0;

let checkPass = function() {
    if (document.getElementById('pass').value !==
        document.getElementById('passx2').value) {
        document.getElementById('passMessage').style.color = 'red';
        document.getElementById('passMessage').innerHTML = 'Пароли не совпадают';
        samePass = false
    } else {
        document.getElementById('passMessage').innerHTML = '';
        samePass = true
    }
}

let checkNickname = function() {
    let elem = document.getElementById('nickname').value
    if (!/^[^@]+$/.test(elem)) {
        document.getElementById('nicknameMessage').style.color = 'red';
        document.getElementById('nicknameMessage').innerHTML = 'Неверный формат nickname';
    } else {
        checkUniqueness("nickname", elem)
    }
}

let checkUniqueness = function(type, val) {
    let xhr = new XMLHttpRequest();
    xhr.open('POST', '/api/exists', true);
    xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
    xhr.onreadystatechange = function() {
        if (this.status !== 200) {
            document.getElementById(type + 'Message').style.color = 'red';
            document.getElementById(type + 'Message').innerHTML = 'Такой ' + type + ' уже существует';
            if (type === 'email') {
                mailUnique = false
            } else {
                nicknameUnique = false
            }
        } else {
            document.getElementById(type + 'Message').innerHTML = '';
            if (type === 'email') {
                mailUnique = true
            } else {
                nicknameUnique = true
            }
        }
    };
    xhr.send('val=' + val + '&type=' + type);
}


let checkEmail = function() {
    let elem = document.getElementById('email').value
    if (!/^([a-z0-9.-]+@[a-z0-9.-]+)$/.test(elem)) {
        document.getElementById('emailMessage').style.color = 'red';
        document.getElementById('emailMessage').innerHTML = 'Неверный формат email';
    } else {
        checkUniqueness("email", elem)
    }
}

let sendMail = function() {
    if (mailUnique) {
        let mail = document.getElementById('email').value
        let xhr = new XMLHttpRequest();
        xhr.open('POST', '/api/sendmail', true);
        xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
        xhr.onreadystatechange = function() {
            if (this.status !== 200) {
                document.getElementById('emailMessage').style.color = 'red';
                document.getElementById('emailMessage').innerHTML = 'Не удалось отправить код';
            } else {
                document.getElementById('emailMessage').style.color = 'green';
                document.getElementById('emailMessage').innerHTML = 'Код отправлен';
            }
        };
        xhr.send('email=' + mail);
    } else {
        document.getElementById('emailMessage').style.color = 'red';
        document.getElementById('emailMessage').innerHTML = 'Такой email уже существует';
    }
}

let checkCode = function() {
    if (mailUnique) {
        let mail = document.getElementById('email').value
        let code = document.getElementById('code').value
        let xhr = new XMLHttpRequest();
        xhr.open('POST', '/api/checkmail', true);
        xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
        xhr.onreadystatechange = function() {
            if (this.status !== 200) {
                document.getElementById('codeMessage').style.color = 'red';
                document.getElementById('codeMessage').innerHTML = 'Неверный код';
            } else {
                document.getElementById('codeMessage').style.color = 'green';
                document.getElementById('codeMessage').innerHTML = 'email подтвержден';
            }
        };
        xhr.send('email=' + mail + '&code=' + code);
    } else {
        document.getElementById('emailMessage').style.color = 'red';
        document.getElementById('emailMessage').innerHTML = 'Неверный формат email';
    }
}

let register = function() {
    if (mailUnique && nicknameUnique && samePass) {
        let mail = document.getElementById('email').value
        let nickname = document.getElementById('nickname').value
        let password = document.getElementById('pass').value
        let xhr = new XMLHttpRequest();
        xhr.open('POST', '/api/register', true);
        xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
        xhr.onreadystatechange = function() {
            if (this.status !== 200) {
                document.getElementById('regMessage').style.color = 'red';
                document.getElementById('regMessage').innerHTML = 'Ошибка при регистрации';
            } else {
                document.getElementById('regMessage').style.color = 'green';
                document.getElementById('regMessage').innerHTML = 'Пользователь зарегистрирован';
                window.open(this.responseText, "_self");
            }
        };
        xhr.send('email=' + mail + '&nickname=' + nickname + '&password=' + password);
    }
}

let login = function() {
    let login = document.getElementById('email').value
    let password = document.getElementById('pass').value
    let xhr = new XMLHttpRequest();
    xhr.open('POST', '/api/login', true);
    xhr.setRequestHeader('Authorization', 'Basic ' + btoa(login + ':' + password));
    xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
    xhr.onreadystatechange = function() {
        if (this.status !== 200) {
            document.getElementById('loginMessage').style.color = 'red';
            document.getElementById('loginMessage').innerHTML = 'Ошибка';
        } else {
            window.open(this.responseText, "_self");
        }
    };
    xhr.send('url=' + window.location.href);
}

let logout = function() {
    let xhr = new XMLHttpRequest();
    xhr.open('POST', '/api/logout', true);
    xhr.onreadystatechange = function() {
        if (this.status === 200) {
            window.open("/", "_self");
        }
    };
    xhr.send();
}

let uploadFile = function() {
    if (document.getElementById('file').value === "") {
        document.getElementById('fileMessage').style.color = 'red';
        document.getElementById('fileMessage').innerHTML = 'Необходимо добавить файл';
    } else {
        document.getElementById('fileMessage').innerHTML = '';
        let filename = document.getElementById('file').files[0].name;
        let bookname = document.getElementById('bookName').value;
        if (bookname === '') {
            document.getElementById('nameMessage').style.color = 'red';
            document.getElementById('nameMessage').innerHTML = 'Необходимо указать название';
            return;
        } else {
            document.getElementById('nameMessage').innerHTML = '';
        }
        let genre = document.getElementById('genre').value;
        if (genre === "-1") {
            genre = "9";
        }
        let authors = [];
        let authorPresent = false;
        for (i = 0; i < curAuthor; i++) {
            let author = document.getElementById('' + i).value;
            if (author !== "") {
                authorPresent = true;
                authors.push(author);
            }
        }
        if (!authorPresent) {
            document.getElementById('authorMessage').style.color = 'red';
            document.getElementById('authorMessage').innerHTML = 'Необходимо указать хотя бы одного автора';
            return;
        } else {
            document.getElementById('authorMessage').innerHTML = '';
        }
        let year = document.getElementById('dateField').value;
        let json = JSON.stringify({"Name": bookname, "File": filename, "Authors": authors, "Genre": genre, "Year": year});

        document.getElementById('fileMessage').innerHTML = '';
        var formData = new FormData(document.forms.upload);
        var xhr = new XMLHttpRequest();
        xhr.open("POST", "/api/files/load");
        xhr.onreadystatechange = function() {
            if (this.status !== 200) {
                document.getElementById('uploadMessage').style.color = 'red';
                document.getElementById('uploadMessage').innerHTML = 'Ошибка';
            } else {
                xhr.open('POST', '/api/addfile', true);
                xhr.onreadystatechange = function () {
                    if (this.status === 200) {
                        document.getElementById('uploadMessage').style.color = 'green';
                        document.getElementById('uploadMessage').innerHTML = 'Файл загружен';
                    } else {
                        document.getElementById('uploadMessage').style.color = 'red';
                        document.getElementById('uploadMessage').innerHTML = 'Ошибка';
                    }
                }
                xhr.send(json);
            }
        };
        xhr.send(formData);
    }
}

let changePass = function() {
    if (samePass) {
        let password = document.getElementById('pass').value
        let xhr = new XMLHttpRequest();
        xhr.open('POST', '/api/change', true);
        xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
        xhr.onreadystatechange = function() {
            if (this.status !== 200) {
                document.getElementById('chPassMessage').style.color = 'red';
                document.getElementById('chPassMessage').innerHTML = 'Ошибка';
            } else {
                document.getElementById('chPassMessage').style.color = 'green';
                document.getElementById('chPassMessage').innerHTML = 'Пароль изменен';
            }
        };
        xhr.send('password=' + password);
    }
}

let find_user = function() {
    let table_div = document.getElementById("table_div");
    let user = document.getElementById('user').value
    let xhr = new XMLHttpRequest();
    xhr.open('POST', '/api/getlist', true);
    xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
    xhr.onreadystatechange = function () {
        table_div.innerHTML = ""
        if (this.status === 200) {
            table = document.createElement("table");
            table.className = "u-table-entity u-sheet"
            table.innerHTML += "<colgroup>\n" +
                "                                <col style=\"width: 33%\">\n" +
                "                                <col style=\"width: 33%\">\n" +
                "                                <col style=\"width: 33%\">\n" +
                "                            </colgroup>\n" +
                "                            <thead class=\"u-align-center u-custom-color-2 u-table-header u-table-header-1\">\n" +
                "                            <tr style=\"height: 66px;\">\n" +
                "                                <th class=\"u-border-3 u-border-white u-table-cell u-table-cell-1\">nickname</th>\n" +
                "                                <th class=\"u-border-3 u-border-white u-table-cell u-table-cell-2\">email</th>\n" +
                "                                <th class=\"u-border-3 u-border-white u-table-cell u-table-cell-3\"></th>\n" +
                "                            </tr>\n" +
                "                            </thead>";
            let jsonResponse = JSON.parse(this.responseText);
            let res = "";
            let i = 0;
            for (let user of jsonResponse.Users) {
                i = i + 1;
                let btn_text = "Забанить";
                if (user.Banned) {
                    btn_text = "Разбанить";
                }
                res += '<tr style="height: 120px;">' +
               ' <td id="name'+ user.Num +'" class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-6">' + user.Nickname +'</td>' +
               ' <td class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-7">' + user.Email +'</td>'+
               ' <td class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-8">' +
               '<a href="#ban"><button type="button" onclick="ban_user(this.id)" id="'+ user.Num +'" class="u-form-group u-form-submit u-border-none u-btn u-btn-submit u-button-style u-custom-color-2 u-text-hover-custom-color-3 u-btn-1">'+ btn_text +'</button></a>' +
               '</td></tr>';
            }
            table.innerHTML += '<tbody class="u-align-center u-table-body u-text-black u-table-body-1">' +
                res + '</tbody>';
            table_div.append(table);
            let div_common = document.createElement('div');
            div_common.style = "display: inline-block; width: 900px";
            if (page !== 0) {
                let div = document.createElement('div');
                div.className = 'u-btn-3';
                div.innerHTML = '<a href="#ban_user"><button type="button" onclick="prev_page_user()" class="u-form-group u-form-submit u-border-none u-btn u-btn-submit u-button-style u-custom-color-2 u-text-hover-custom-color-3 u-btn-1">Предыдущая</button></a>';
                div_common.append(div);
            }
            if (i === 10) {
                let div = document.createElement('div');
                div.className = 'u-btn-4';
                div.innerHTML = '<a href="#ban_user"><button type="button" onclick="next_page_user()" class="u-form-group u-form-submit u-border-none u-btn u-btn-submit u-button-style u-custom-color-2 u-text-hover-custom-color-3 u-btn-1">Следующая</button></a>';
                div_common.append(div);
            }
            table_div.append(div_common);
        }
    }
    xhr.send('type=user' + '&user=' + user + '&page=' + page);
}

let find_user_var = function() {
    let list = document.getElementById("users");
    let user = document.getElementById('user').value;
    let xhr = new XMLHttpRequest();
    xhr.open('POST', '/api/getlist', true);
    xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
    xhr.onreadystatechange = function () {
        list.innerHTML = "";
        if (this.status === 200) {
            let jsonResponse = JSON.parse(this.responseText);
            for (let user of jsonResponse.Users) {
                list.innerHTML += "<option>"+ user.Nickname +"</option>"
            }
        }
    }
    xhr.send('type=user' + '&user=' + user + '&page=0');
}

let ban_user = function(id) {
    document.getElementById('banMessage').innerHTML = "";
    document.getElementById("ban_reason").value = "";
    document.getElementById("username").innerText = document.getElementById("name" + id).innerText;
    document.getElementById("ban_btn").innerText = document.getElementById(id).innerText;
}

let ban = function() {
    let user = document.getElementById("username").innerText;
    let msg = document.getElementById("ban_reason").value;
    if (msg !== "") {
        let xhr = new XMLHttpRequest();
        xhr.open('POST', '/api/ban', true);
        xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
        xhr.onreadystatechange = function () {
            if (this.status === 200) {
                document.getElementById('banMessage').style.color = 'green';
                document.getElementById('banMessage').innerHTML = 'Статус пользователя изменен';
            } else {
                document.getElementById('banMessage').style.color = 'red';
                document.getElementById('banMessage').innerHTML = 'Ошибка';
            }
        }
        xhr.send('user=' + user + '&msg=' + msg);
    } else {
        document.getElementById('banMessage').style.color = 'red';
        document.getElementById('banMessage').innerHTML = 'Необходимо ввести причину';
    }
}

let setMenu = function() {
    let menu = document.getElementById("tabs");
    let xhr = new XMLHttpRequest();
    xhr.open('POST', '/api/getrights', true);
    xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
    xhr.onreadystatechange = function () {
        if (this.status === 200) {
            if (this.responseText === "10") {
                menu.innerHTML = '<a href="#check_book" onclick="clear_admin_reader(); setChecks()">Проверить книгу</a>' +
                '<a href="#change_pass" onclick="clear_admin_reader()">Сменить пароль</a>';
            } else if (this.responseText === "01") {
                menu.innerHTML = '<a href="#ban_user" onclick="clear_admin_reader()">Забанить пользователя</a>' +
                '<a href="#change_pass" onclick="clear_admin_reader()">Сменить пароль</a>'
            } else {
                menu.innerHTML = '<a href="#ban_user" onclick="clear_admin_reader()">Забанить пользователя</a>\n' +
                    '                <a href="#check_book" onclick="clear_admin_reader(); setChecks()">Проверить книгу</a>\n' +
                    '                <a href="#change_pass" onclick="clear_admin_reader()">Сменить пароль</a>'
            }
        }
    }
    xhr.send('url=' + window.location.href);
}

let find_author = function(id) {
    let list = document.getElementById('authorlist' + id);
    let author = document.getElementById(id).value;
    let xhr = new XMLHttpRequest();
    xhr.open('POST', '/api/getlist', true);
    xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
    xhr.onreadystatechange = function () {
        list.innerHTML = "";
        if (this.status === 200) {
            let jsonResponse = JSON.parse(this.responseText);
            for (let author of jsonResponse.Authors) {
                list.innerHTML += "<option>"+ author +"</option>"
            }
        }
    }
    xhr.send('type=author' + '&author=' + author);
}

let find_names_user = function() {
    let list = document.getElementById('names');
    let name = document.getElementById("bookName_find").value;
    let xhr = new XMLHttpRequest();
    xhr.open('POST', '/api/getlist', true);
    xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
    xhr.onreadystatechange = function () {
        list.innerHTML = "";
        if (this.status === 200) {
            let jsonResponse = JSON.parse(this.responseText);
            for (let name of jsonResponse.Entries) {
                list.innerHTML += "<option>"+ name +"</option>"
            }
        }
    }
    xhr.send('type=book_name_user' + '&name=' + name);
}

let find_owners = function() {
    let list = document.getElementById('owners');
    let owner = document.getElementById("owner").value;
    let xhr = new XMLHttpRequest();
    xhr.open('POST', '/api/getlist', true);
    xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
    xhr.onreadystatechange = function () {
        list.innerHTML = "";
        if (this.status === 200) {
            let jsonResponse = JSON.parse(this.responseText);
            for (let owner of jsonResponse.Entries) {
                list.innerHTML += "<option>"+ owner +"</option>"
            }
        }
    }
    xhr.send('type=owner' + '&owner=' + owner);
}

let find_names = function() {
    let list = document.getElementById('names');
    let name = document.getElementById("bookName_find").value;
    let xhr = new XMLHttpRequest();
    xhr.open('POST', '/api/getlist', true);
    xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
    xhr.onreadystatechange = function () {
        list.innerHTML = "";
        if (this.status === 200) {
            let jsonResponse = JSON.parse(this.responseText);
            for (let name of jsonResponse.Entries) {
                list.innerHTML += "<option>"+ name +"</option>"
            }
        }
    }
    xhr.send('type=book_name_gen' + '&name=' + name);
}

let curAuthor = 1;

let addAuthor = function() {
    let authors = document.getElementById("authors");
    let span = document.createElement("span");
    span.innerHTML = '<input placeholder="Автор" id="' + curAuthor + '" onInput="find_author(this.id)" ' +
    'list="authorlist' + curAuthor + '" class="u-border-1 u-border-grey-30 u-input u-input-rectangle u-input-1"/>' +
    '<datalist id="authorlist' + curAuthor + '"></datalist>';
    authors.append(span);
    curAuthor += 1;
}

let next_page_reader = function(id, type) {
    source_page += 1;
    get_book_data(id, type);
}

let prev_page_reader = function(id, type) {
    source_page -= 1;
    get_book_data(id, type);
}

let next_page_user = function() {
    page += 1;
    find_user();
}

let prev_page_user = function() {
    page -= 1;
    find_user();
}

let next_page_book = function() {
    book_page += 1;
    setChecks();
}

let prev_page_book = function() {
    book_page -= 1;
    setChecks();
}

let next_user_page_book = function() {
    user_book_page += 1;
    findBook();
}

let prev_user_page_book = function() {
    user_book_page -= 1;
    findBook();
}

let next_gen_page_book = function() {
    gen_book_page += 1;
    findBook_gen();
}

let prev_gen_page_book = function() {
    gen_book_page -= 1;
    findBook_gen();
}

let setChecks = function() {
    let table_div = document.getElementById("table_div_check");
    let xhr = new XMLHttpRequest();
    xhr.open('POST', '/api/getlist', true);
    xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
    xhr.onreadystatechange = function () {
        table_div.innerHTML = ""
        if (this.status === 200) {
            table = document.createElement("table");
            table.className = "u-table-entity u-sheet"
            table.innerHTML += "<colgroup>\n" +
                "                                <col style=\"width: 67%\">\n" +
                "                                <col style=\"width: 33%\">\n" +
                "                            </colgroup>\n" +
                "                            <thead class=\"u-align-center u-custom-color-2 u-table-header u-table-header-1\">\n" +
                "                            <tr style=\"height: 66px;\">\n" +
                "                                <th class=\"u-border-3 u-border-white u-table-cell u-table-cell-1\">Название</th>\n" +
                "                                <th class=\"u-border-3 u-border-white u-table-cell u-table-cell-2\"></th>\n" +
                "                            </tr>\n" +
                "                            </thead>";
            let jsonResponse = JSON.parse(this.responseText);
            let res = "";
            let i = 0;
            for (let book of jsonResponse.Books) {
                i = i + 1;
                res += '<tr style="height: 120px;">' +
                    ' <td id="book'+ book.Guid +'" class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-6">' + book.Name +'</td>' +
                    ' <td class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-7">' +
                    '<a href="#checker"><button type="button" onclick="check_book(this.id)" id="'+ book.Guid +'" class="u-form-group u-form-submit u-border-none u-btn u-btn-submit u-button-style u-custom-color-2 u-text-hover-custom-color-3 u-btn-1">Проверить</button></a>' +
                    '</td></tr>';
            }
            table.innerHTML += '<tbody class="u-align-center u-table-body u-text-black u-table-body-1">' +
                res + '</tbody>';
            table_div.append(table);
            let div_common = document.createElement('div');
            div_common.style = "display: inline-block; width: 900px";
            if (book_page !== 0) {
                let div = document.createElement('div');
                div.className = 'u-btn-3';
                div.innerHTML = '<a href="#check_book"><button type="button" onclick="prev_page_book()" class="u-form-group u-form-submit u-border-none u-btn u-btn-submit u-button-style u-custom-color-2 u-text-hover-custom-color-3 u-btn-1">Предыдущая</button></a>';
                div_common.append(div);
            }
            if (i === 10) {
                let div = document.createElement('div');
                div.className = 'u-btn-4';
                div.innerHTML = '<a href="#check_book"><button type="button" onclick="next_page_book()" class="u-form-group u-form-submit u-border-none u-btn u-btn-submit u-button-style u-custom-color-2 u-text-hover-custom-color-3 u-btn-1">Следующая</button></a>';
                div_common.append(div);
            }
            table_div.append(div_common);
        }
    }
    xhr.send('type=admin_book' + '&page=' + book_page);
}

let get_genre = function(code)  {
    switch (code) {
        case "0":
            return "Фантастика"
        case "1":
            return "Фэнтези"
        case "2":
            return "Комиксы"
        case "3":
            return "Сатира"
        case "4":
            return "Детектив"
        case "5":
            return "Приключения"
        case "6":
            return "История"
        case "7":
            return "Религия"
        case "8":
            return "Ужасы"
        default:
            return "Прочее"
    }
}

let check_book = function(id) {
    let main_div = document.getElementById('checker');
    let reader_emb = document.getElementById("reader_emb")
    let xhr = new XMLHttpRequest();
    xhr.open('POST', '/api/getbookmeta', true);
    xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
    xhr.onreadystatechange = function() {
        if (this.status === 200) {
            main_div.innerHTML = ""
            reader_emb.innerHTML = ""
            let jsonResponse = JSON.parse(this.responseText);
            book_page_num = jsonResponse.Pagenum;
            document.getElementById("ownername").innerText = jsonResponse.Nickname;
            document.getElementById("bookname").innerText = jsonResponse.Name;
            document.getElementById("decl_btn").innerHTML = '<button type="button" onclick="decline(\'' + id + '\')" class="u-form-group u-form-submit u-border-none u-btn u-btn-submit u-button-style u-custom-color-2 u-text-hover-custom-color-3 u-btn-1">Отклонить</button>'
            let div = document.createElement("div");
            div.innerText = jsonResponse.Name;
            div.style = "text-align: center; font-size: 2rem; color: #727272"
            main_div.append(div);
            let table_div = document.createElement('div');
            table_div.className = "u-table u-table-1"
            table = document.createElement("table");
            table.className = "u-table-entity u-sheet"
            table.innerHTML += "<colgroup>\n" +
                "                                <col style=\"width: 25%\">\n" +
                "                                <col style=\"width: 25%\">\n" +
                "                                <col style=\"width: 25%\">\n" +
                "                                <col style=\"width: 25%\">\n" +
                "                            </colgroup>\n"
            let authors = "";
            for (let author of jsonResponse.Authors) {
                authors += author + ", ";
            }
            authors = authors.substr(0, authors.length - 2);
            table.innerHTML += '<tr style="height: 120px;">' +
                ' <td style="font-size: 1.2rem; color: #514f4f;" class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-6">Владелец</td>' +
                ' <td class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-7">' + jsonResponse.Nickname + '</td>' +
                ' <td style="font-size: 1.2rem; color: #514f4f;" class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-8">Жанр</td>' +
                ' <td class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-9">' + get_genre(jsonResponse.Genre) + '</td></tr>' +
                '<tr style="height: 120px;"><td style="font-size: 1.2rem; color: #514f4f;" class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-6">Год написания</td>' +
                ' <td class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-7">' + jsonResponse.Year + '</td>' +
                ' <td style="font-size: 1.2rem; color: #514f4f;" class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-8">Авторы</td>' +
                ' <td class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-9">' + authors + '</td></tr>';
            table_div.append(table)
            main_div.append(table_div)
            let div_common = document.createElement('div');
            div_common.style = "display: inline-block; width: 900px";
            let div1 = document.createElement('div');
            div1.className = 'u-btn-3';
            div1.innerHTML = '<a href="#check_book" onclick="clear_admin_reader()"><button type="button" onclick="approve(\'' + id + '\')" class="u-form-group u-form-submit u-border-none u-btn u-btn-submit u-button-style u-custom-color-2 u-text-hover-custom-color-3 u-btn-1">Одобрить</button></a>';
            div_common.append(div1);
            let div2 = document.createElement('div');
            div2.className = 'u-btn-4';
            div2.innerHTML = '<a href="#check_res"><button type="button" class="u-form-group u-form-submit u-border-none u-btn u-btn-submit u-button-style u-custom-color-2 u-text-hover-custom-color-3 u-btn-1">Отклонить</button></a>';
            div_common.append(div2);
            main_div.append(div_common);
            let reader = document.createElement("div")
            reader.id = "reader"
            reader_emb.append(reader)
            reader_emb.style = "min-height: 910px; margin-top: 100px"
            get_book_data(id, "admin")
        }
    }
    xhr.send('type=check' + '&guid=' + id);
}

let approve = function(id) {
    let xhr = new XMLHttpRequest();
    xhr.open('POST', '/api/approve', true);
    xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
    xhr.onreadystatechange = function () {
        if (this.status === 200) {
            document.getElementById('dataMessage').style.color = 'green';
            document.getElementById('dataMessage').innerHTML = 'Статус книги изменен';
        } else {
            document.getElementById('dataMessage').style.color = 'red';
            document.getElementById('dataMessage').innerHTML = 'Ошибка';
        }
    }
    xhr.send('guid=' + id);
}

let decline = function(id) {
    let user = document.getElementById("ownername").innerText;
    let msg = document.getElementById("decline_reason").value;
    if (msg !== "") {
        let xhr = new XMLHttpRequest();
        xhr.open('POST', '/api/decline', true);
        xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
        xhr.onreadystatechange = function () {
            if (this.status === 200) {
                document.getElementById('dataMessage').style.color = 'green';
                document.getElementById('dataMessage').innerHTML = 'Статус книги изменен';
            } else {
                document.getElementById('dataMessage').style.color = 'red';
                document.getElementById('dataMessage').innerHTML = 'Ошибка';
            }
        }
        xhr.send('guid=' + id + '&user=' + user + '&msg=' + msg);
    } else {
        document.getElementById('dataMessage').style.color = 'red';
        document.getElementById('dataMessage').innerHTML = 'Необходимо ввести причину';
    }
}

let get_book_data = function(id, type) {
    let main_div = document.getElementById("reader");
    let xhr = new XMLHttpRequest();
    xhr.open('POST', '/api/getbookpage', true);
    xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
    xhr.responseType = "blob";
    xhr.onreadystatechange = function () {
        main_div.innerHTML = "";
        if (this.status === 200) {
            document.getElementById('dataMessage').innerHTML = '';
            let embed = document.createElement("embed");
            embed.className = "embed";
            embed.type = "application/pdf";
            embed.width = 900;
            embed.height = 910;
            embed.src = URL.createObjectURL(this.response);
            main_div.append(embed)
            if (source_page !== 0) {
                let div = document.createElement('div');
                div.className = 'u-btn-6';
                div.innerHTML = '<button type="button" onclick="prev_page_reader(\'' + id + '\',\'' + type + '\')" class="u-form-group u-form-submit u-border-none u-btn u-btn-submit u-button-style u-custom-color-2 u-text-hover-custom-color-3 u-btn-1">Предыдущая</button>';
                main_div.append(div);
            }
            if (source_page < book_page_num - 1) {
                let div = document.createElement('div');
                div.className = 'u-btn-7';
                div.innerHTML = '<button type="button" onclick="next_page_reader(\'' + id + '\',\'' + type + '\')" class="u-form-group u-form-submit u-border-none u-btn u-btn-submit u-button-style u-custom-color-2 u-text-hover-custom-color-3 u-btn-1">Следующая</button>';
                main_div.append(div);
            }
        } else {
            document.getElementById('dataMessage').style.color = 'red';
            document.getElementById('dataMessage').innerHTML = 'Ошибка';
        }
    }
    xhr.send('type=' + type + '&guid=' + id + '&page=' + source_page)
}

let get_book = function(id) {
    source_page = 0;
    let main_div = document.getElementById('book_reader');
    let reader_emb = document.getElementById("reader_emb")
    let xhr = new XMLHttpRequest();
    xhr.open('POST', '/api/getbookmeta', true);
    xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
    xhr.onreadystatechange = function () {
        if (this.status === 200) {
            main_div.innerHTML = ""
            reader_emb.innerHTML = ""
            let jsonResponse = JSON.parse(this.responseText);
            book_page_num = jsonResponse.Pagenum;
            let div = document.createElement("div");
            div.innerText = jsonResponse.Name;
            div.style = "text-align: center; font-size: 2rem; color: #727272"
            main_div.append(div);
            let table_div = document.createElement('div');
            table_div.className = "u-table u-table-1"
            table = document.createElement("table");
            table.className = "u-table-entity u-sheet"
            table.innerHTML += "<colgroup>\n" +
                "                                <col style=\"width: 25%\">\n" +
                "                                <col style=\"width: 25%\">\n" +
                "                                <col style=\"width: 25%\">\n" +
                "                                <col style=\"width: 25%\">\n" +
                "                            </colgroup>\n"
            let authors = "";
            for (let author of jsonResponse.Authors) {
                authors += author + ", ";
            }
            authors = authors.substr(0, authors.length - 2);
            table.innerHTML += '<tr style="height: 120px;">' +
                ' <td style="font-size: 1.2rem; color: #514f4f;" class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-6">Владелец</td>' +
                ' <td class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-7">' + jsonResponse.Nickname + '</td>' +
                ' <td style="font-size: 1.2rem; color: #514f4f;" class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-8">Жанр</td>' +
                ' <td class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-9">' + get_genre(jsonResponse.Genre) + '</td></tr>' +
                ' <tr style="height: 120px;"><td style="font-size: 1.2rem; color: #514f4f;" class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-6">Год написания</td>' +
                ' <td class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-7">' + jsonResponse.Year + '</td>' +
                ' <td style="font-size: 1.2rem; color: #514f4f;" class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-8">Авторы</td>' +
                ' <td class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-9">' + authors + '</td></tr>';
            table_div.append(table)
            main_div.append(table_div)
            let reader = document.createElement("div")
            reader.id = "reader"
            reader_emb.append(reader)
            reader_emb.style = "min-height: 910px;"
            get_book_data(id, "gen")
            location.href = "#book_reader"
        }
    }
    xhr.send('type=meta' + '&guid=' + id);
}

let my_book = function(id) {
    source_page = 0;
    let main_div = document.getElementById('user_book_reader');
    let reader_emb = document.getElementById("reader_emb")
    let xhr = new XMLHttpRequest();
    xhr.open('POST', '/api/getbookmeta', true);
    xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
    xhr.onreadystatechange = function() {
        if (this.status === 200) {
            main_div.innerHTML = ""
            reader_emb.innerHTML = ""
            let jsonResponse = JSON.parse(this.responseText);
            book_page_num = jsonResponse.Pagenum;
            let div = document.createElement("div");
            div.innerText = jsonResponse.Name;
            div.style = "text-align: center; font-size: 2rem; color: #727272"
            main_div.append(div);
            let table_div = document.createElement('div');
            table_div.className = "u-table u-table-1"
            table = document.createElement("table");
            table.className = "u-table-entity u-sheet"
            table.innerHTML += "<colgroup>\n" +
                "                                <col style=\"width: 25%\">\n" +
                "                                <col style=\"width: 25%\">\n" +
                "                                <col style=\"width: 25%\">\n" +
                "                                <col style=\"width: 25%\">\n" +
                "                            </colgroup>\n"
            let authors = "";
            for (let author of jsonResponse.Authors) {
                authors += author + ", ";
            }
            authors = authors.substr(0, authors.length - 2);
            table.innerHTML += '<tr style="height: 120px;">' +
                ' <td style="font-size: 1.2rem; color: #514f4f;" class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-6">Жанр</td>' +
                ' <td class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-7">' + get_genre(jsonResponse.Genre) + '</td>' +
                ' <td style="font-size: 1.2rem; color: #514f4f;" class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-8">Год написания</td>' +
                ' <td class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-9">' + jsonResponse.Year + '</td></tr>' +
                ' <tr style="height: 120px;"><td style="font-size: 1.2rem; color: #514f4f;" class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-8">Авторы</td>' +
                ' <td class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-9">' + authors + '</td></tr>';
            table_div.append(table)
            main_div.append(table_div)
            let reader = document.createElement("div")
            reader.id = "reader"
            reader_emb.append(reader)
            reader_emb.style = "min-height: 910px;"
            get_book_data(id, "user")
        }
    }
    xhr.send('type=self' + '&guid=' + id);
}

let request_source = function() {
    let xhr = new XMLHttpRequest();
    xhr.open('POST', '/api/getbookdata', true);
    xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
    xhr.onreadystatechange = function() {

    }
    xhr.send()
}

let setAuthors = function() {
    document.getElementById("authors_find").innerHTML = ""
    document.getElementById("authors_load").innerHTML = '<span><p id="authors">' +
                                        '<input placeholder="Автор" id="0" oninput="find_author(this.id)" list="authorlist0" class="u-border-1 u-border-grey-30 u-input u-input-rectangle u-input-1"/>' +
                                        '<datalist id="authorlist0">' +
                                        '</datalist>' +
                                    '</p></span>'
    document.getElementById("user_book_reader").innerHTML = "";
}

let setAuthors_find = function() {
    document.getElementById("authors_load").innerHTML = ""
    document.getElementById("authors_find").innerHTML = '<span><p id="authors">' +
                                        '<input placeholder="Автор" id="0" oninput="find_author(this.id)" list="authorlist0" class="u-border-1 u-border-grey-30 u-input u-input-rectangle u-input-1"/>' +
                                        '<datalist id="authorlist0">' +
                                        '</datalist>' +
                                    '</p></span>'
    document.getElementById("user_book_reader").innerHTML = "";
}

let findBook = function() {
    let bookname = document.getElementById('bookName_find').value;
    let genre = document.getElementById('genre_find').value;
    if (genre === "-1") {
        genre = "";
    }
    let authors = [];
    for (i = 0; i < curAuthor; i++) {
        let author = document.getElementById('' + i).value;
        if (author !== "") {
            authors.push(author);
        }
    }
    let maxyear = document.getElementById('dateFieldNotAfter_find').value;
    let minyear = document.getElementById('dateFieldNotBefore_find').value;
    let json = JSON.stringify({"Name": bookname, "Nickname": "", "Authors": authors, "Genre": genre,
        "MaxYear": maxyear, "MinYear": minyear, "Page": user_book_page});
    let table_div = document.getElementById("user_book_list");
    let xhr = new XMLHttpRequest();
    xhr.open('POST', '/api/mybooks', true);
    xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
    xhr.onreadystatechange = function () {
        table_div.innerHTML = ""
        if (this.status === 200) {
            table = document.createElement("table");
            table.className = "u-table-entity u-sheet"
            table.innerHTML += "<colgroup>\n" +
                "                                <col style=\"width: 20%\">\n" +
                "                                <col style=\"width: 20%\">\n" +
                "                                <col style=\"width: 20%\">\n" +
                "                                <col style=\"width: 20%\">\n" +
                "                            </colgroup>\n" +
                "                            <thead class=\"u-align-center u-custom-color-2 u-table-header u-table-header-1\">\n" +
                "                            <tr style=\"height: 66px;\">\n" +
                "                                <th class=\"u-border-3 u-border-white u-table-cell u-table-cell-1\">Название</th>\n" +
                "                                <th class=\"u-border-3 u-border-white u-table-cell u-table-cell-1\">Жанр</th>\n" +
                "                                <th class=\"u-border-3 u-border-white u-table-cell u-table-cell-1\">Год написания</th>\n" +
                "                                <th class=\"u-border-3 u-border-white u-table-cell u-table-cell-1\">Авторы</th>\n" +
                "                                <th class=\"u-border-3 u-border-white u-table-cell u-table-cell-2\"></th>\n" +
                "                            </tr>\n" +
                "                            </thead>";
            let jsonResponse = JSON.parse(this.responseText);
            let res = "";
            let i = 0;
            for (let book of jsonResponse.Books) {
                i = i + 1;
                let authors = "";
                for (let author of book.Authors) {
                    authors += author + ", ";
                }
                authors = authors.substr(0, authors.length - 2);
                res += '<tr style="height: 120px;">' +
                    ' <td id="book'+ book.Guid +'" class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-6">' + book.Name +'</td>' +
                    ' <td class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-6">' + get_genre(book.Genre) +'</td>' +
                    ' <td class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-6">' + book.Year +'</td>' +
                    ' <td class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-6">' + authors +'</td>' +
                    ' <td class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-7">' +
                    '<a href="#user_book_reader"><button type="button" onclick="my_book(this.id)" id="'+ book.Guid +'" class="u-form-group u-form-submit u-border-none u-btn u-btn-submit u-button-style u-custom-color-2 u-text-hover-custom-color-3 u-btn-1">Читать</button></a>' +
                    '</td></tr>';
            }
            table.innerHTML += '<tbody class="u-align-center u-table-body u-text-black u-table-body-1">' +
                res + '</tbody>';
            table_div.append(table);
            let div_common = document.createElement('div');
            div_common.style = "display: inline-block; width: 900px";
            if (user_book_page !== 0) {
                let div = document.createElement('div');
                div.className = 'u-btn-3';
                div.innerHTML = '<a href="#my_books"><button type="button" onclick="prev_user_page_book()" class="u-form-group u-form-submit u-border-none u-btn u-btn-submit u-button-style u-custom-color-2 u-text-hover-custom-color-3 u-btn-1">Предыдущая</button></a>';
                div_common.append(div);
            }
            if (i === 10) {
                let div = document.createElement('div');
                div.className = 'u-btn-4';
                div.innerHTML = '<a href="#my_books"><button type="button" onclick="next_user_page_book()" class="u-form-group u-form-submit u-border-none u-btn u-btn-submit u-button-style u-custom-color-2 u-text-hover-custom-color-3 u-btn-1">Следующая</button></a>';
                div_common.append(div);
            }
            table_div.append(div_common);
        }
    }
    xhr.send(json);
}

let findBook_gen = function() {
    let owner = document.getElementById('owner').value;
    let bookname = document.getElementById('bookName_find').value;
    let genre = document.getElementById('genre_find').value;
    if (genre === "-1") {
        genre = "";
    }
    let authors = [];
    for (i = 0; i < curAuthor; i++) {
        let author = document.getElementById('' + i).value;
        if (author !== "") {
            authors.push(author);
        }
    }
    let maxyear = document.getElementById('dateFieldNotAfter_find').value;
    let minyear = document.getElementById('dateFieldNotBefore_find').value;
    let json = JSON.stringify({"Name": bookname, "Nickname": owner, "Authors": authors, "Genre": genre,
        "MaxYear": maxyear, "MinYear": minyear, "Page": gen_book_page});
    let table_div = document.getElementById("book_list");
    let xhr = new XMLHttpRequest();
    xhr.open('POST', '/api/books', true);
    xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
    xhr.onreadystatechange = function () {
        table_div.innerHTML = ""
        if (this.status === 200) {
            table = document.createElement("table");
            table.className = "u-table-entity u-sheet"
            table.innerHTML += "<colgroup>\n" +
                "                                <col style=\"width: 16%\">\n" +
                "                                <col style=\"width: 16%\">\n" +
                "                                <col style=\"width: 16%\">\n" +
                "                                <col style=\"width: 16%\">\n" +
                "                                <col style=\"width: 16%\">\n" +
                "                            </colgroup>\n" +
                "                            <thead class=\"u-align-center u-custom-color-2 u-table-header u-table-header-1\">\n" +
                "                            <tr style=\"height: 66px;\">\n" +
                "                                <th class=\"u-border-3 u-border-white u-table-cell u-table-cell-1\">Название</th>\n" +
                "                                <th class=\"u-border-3 u-border-white u-table-cell u-table-cell-1\">Жанр</th>\n" +
                "                                <th class=\"u-border-3 u-border-white u-table-cell u-table-cell-1\">Год написания</th>\n" +
                "                                <th class=\"u-border-3 u-border-white u-table-cell u-table-cell-1\">Авторы</th>\n" +
                "                                <th class=\"u-border-3 u-border-white u-table-cell u-table-cell-1\">Владелец</th>\n" +
                "                                <th class=\"u-border-3 u-border-white u-table-cell u-table-cell-2\"></th>\n" +
                "                            </tr>\n" +
                "                            </thead>";
            let jsonResponse = JSON.parse(this.responseText);
            let res = "";
            let i = 0;
            for (let book of jsonResponse.Books) {
                i = i + 1;
                let authors = "";
                for (let author of book.Authors) {
                    authors += author + ", ";
                }
                authors = authors.substr(0, authors.length - 2);
                res += '<tr style="height: 120px;">' +
                    ' <td id="book'+ book.Guid +'" class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-6">' + book.Name +'</td>' +
                    ' <td class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-6">' + get_genre(book.Genre) +'</td>' +
                    ' <td class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-6">' + book.Year +'</td>' +
                    ' <td class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-6">' + authors +'</td>' +
                    ' <td class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-6">' + book.Nickname +'</td>' +
                    ' <td class="u-border-1 u-border-black u-border-no-left u-border-no-right u-table-cell u-table-cell-7">' +
                    ' <button type="button" onclick="get_book(this.id)" id="'+ book.Guid +'" class="u-form-group u-form-submit u-border-none u-btn u-btn-submit u-button-style u-custom-color-2 u-text-hover-custom-color-3 u-btn-1">Читать</button>' +
                    '</td></tr>';
            }
            table.innerHTML += '<tbody class="u-align-center u-table-body u-text-black u-table-body-1">' +
                res + '</tbody>';
            table_div.append(table);
            let div_common = document.createElement('div');
            div_common.style = "display: inline-block; width: 1000px";
            if (gen_book_page !== 0) {
                let div = document.createElement('div');
                div.className = 'u-btn-3';
                div.innerHTML = '<a href="#book_list"><button type="button" onclick="prev_gen_page_book()" class="u-form-group u-form-submit u-border-none u-btn u-btn-submit u-button-style u-custom-color-2 u-text-hover-custom-color-3 u-btn-1">Предыдущая</button></a>';
                div_common.append(div);
            }
            if (i === 10) {
                let div = document.createElement('div');
                div.className = 'u-btn-5';
                div.innerHTML = '<a href="#book_list"><button type="button" onclick="next_gen_page_book()" class="u-form-group u-form-submit u-border-none u-btn u-btn-submit u-button-style u-custom-color-2 u-text-hover-custom-color-3 u-btn-1">Следующая</button></a>';
                div_common.append(div);
            }
            table_div.append(div_common);
        }
    }
    xhr.send(json);
}

let clear_reader = function() {
    document.getElementById("user_book_reader").innerHTML = "";
}

let clear_admin_reader = function() {
    document.getElementById("checker").innerHTML = "";
}