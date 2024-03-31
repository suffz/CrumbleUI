var current = ""
let accounts = [];

function SwapGames(NewGame) {
    if (current != "") {
        document.getElementById(current).style.display = "none";
    }
    document.getElementById(NewGame).style.display = "grid";
    current = NewGame
};

const makePopupHTML = (content, extra, action, action_) => {
    id = new Date().getTime();

    return (html = `
<div class="modal modal-open"id="${id}">
    <div class="modal-box">${content}
        ${extra}
        <div class="modal-action">${action.replace("{id}", id)} ${action_.replace("{id}", id)}</div>
    </div>
</div>`);
};

const closePopup = (id, checked, result_yes, result_no) => {
    const popup = document.getElementById(id);
    popup.classList.remove("modal-open");

    popup.remove()

    if (checked === "Yes") {
        result_yes()
    } else if (checked === "No") {
        result_no()
    }

};

const showPopup = (html) => {
    document.body.insertAdjacentHTML("beforeend", html);
};

const popInfo = (content, extra, result_yes, result_no) => {
    const action = `<button onclick="closePopup('{id}', 'Yes', ${result_yes}, ${result_no})" for="my-modal-2" class="btn">Yes</button>`;
    const action_no = `<button onclick="closePopup('{id}', 'No', ${result_yes}, ${result_no})" for="my-modal-2" class="btn">No</button>`;
    const popup = makePopupHTML(content, extra, action, action_no);
    showPopup(popup);
};

function modalOpen(id, event) {
    let modal = document.getElementById(id);
    modal.classList.add(event);
}

function modalClose(id, event) {
    let modal = document.getElementById(id);
    modal.classList.remove(event);
}

const formatTime = (time) => {
    time = parseInt((time + ""));
    if (!time) {
        return "-";
    }

    let date = new Date(time * 1000);

    // this will break in 900 years
    if (date.getUTCFullYear() > 3000) {
        date = new Date(time);
    }

    return date.toLocaleString().slice(0, -3) + "." + String(date.getMilliseconds());
    // return `${date.getHours()}:${date.getMinutes()}:${date.getSeconds()}`
}

const get_vps_html = (acc) => {
    let color = "red";
    if (acc.status == "Online") {
        color = "green";
    }

    return `
    <tr>
    <td>${acc.ip}</td> 
    <td>${acc.host}</td> 
    <td>
        <span class="text-${color}-500">
            ${acc.status}
        </span>
    </td>
  </tr>
  `;
};

const add_vps_html = (acc) => {
    let html = get_vps_html(acc);
    document.getElementById("vps_list").innerHTML += html;
};

const get_account_html = (acc) => {
    let status_color = "red";
    let usable = "No";
    if (acc.AccountType != "Pending..." && acc.AccountType) {
        status_color = "green";
        usable = "Yes";
    }

    return `
    <tr>
    <td>${acc.Email}</td> 
    <td>${acc.AccountType || "None"}</td> 
    <td>
        <span class="text-${status_color}-500">
            ${usable}
        </span>
    </td>
  </tr>
  `;
};

const add_account_html = (acc) => {
    let html = get_account_html(acc);
    document.getElementById("accounts_list").innerHTML += html;
};

const send_session = () => {
    const IP = document.getElementById("vps_ip").value.trim();
    const Password = document.getElementById("vps_password").value.trim();
    const Host = document.getElementById("vps_user").value.trim();
    socket.send(JSON.stringify(
        {
            type: "add_session",
            task: {
                session: {
                    ip: IP,
                    port: "22",
                    password: Password,
                    host: Host,
                },
            },
        }
    ));
}

const mass_add_accounts_handler = () => {
    const lines = document.getElementById("mass_accounts").value;
    const tmpAccs = lines.split("\n");
    var acc_sub = []
    if (lines.length != 0) {
        tmpAccs.forEach((acc) => {
            if (acc.split(":")[0]) {
                if (acc.split(":")[1]) {
                    acc_sub.push({ email: acc.split(":")[0], password: acc.split(":")[1] })
                }
            }
        });
    }
    window.go.main.App.AuthAccounts(acc_sub).then(resp => {
        resp.forEach((acc) => {
            let found
            let data = document.getElementById("accounts_list").children;
            for (let i = 0; i < data.length; i++) {
                if (data[i].children[0].innerText == acc.Email) {
                    found = true;
                }
            }
            if (!found) {
                add_account_html(acc);
            }
        })
    })
};

const mass_add_accounts_handler_proxys = () => {
    const lines = document.getElementById("mass_accounts_proxys").value;
    const tmpAccs = lines.split("\n");
    var acc_sub = []
    if (lines.length != 0) {
        tmpAccs.forEach((acc) => {
            if (acc != "") {
                let data = acc.split(":");
                if (data.length > 2) {
                    acc_sub.push({
                        Ip: data[0],
                        Port: data[1],
                        User: data[2],
                        Password: data[3],
                    })
                } else {
                    acc_sub.push({
                        Ip: data[0],
                        Port: data[1],
                    })
                }
            }

        });
    }
    window.go.main.App.AddProxys(acc_sub).then(resp => {
        let data = ""
        resp.forEach((acc) => {
            if (acc.Password) {
                hash = ""
                for (let i = 0; i < acc.Password.length; i++) {
                    hash += "."
                }
                acc.Password = hash
            }

            data += (`<tr>
            <td>
                <span class="font-mono">
                    ${acc.IP}
                </span>    
            </td> 
            <td>
                <span class="font-mono">
                    ${acc.Port}
                </span>    
            </td>
            <td>
                <span class="font-mono">
                    ${acc.User}
                </span>    
            </td>
            <td>
                <span class="font-mono">
                    ${acc.Password}
                </span>    
            </td>
          </tr>`)
        })

        document.getElementById("task_list_proxy").innerHTML = data
    })
};

function add_all_proxys() {
    window.go.main.App.GetProxys().then(resp => {
        let all = ""
        resp.forEach((acc) => {
            all += acc + "\n"
        })

        document.getElementById("mass_accounts_proxys").value = all
    })
}

function getLogs() {

}

function contains(arr) {
    for (var i = 0; i < arr.length; i++) {
        if (String(arr[i]).split(":")[2] === "200") return true;
    }
    return false;
}

const add_logs = (acc) => {
    window.go.main.App.GetTaskLogs(acc).then(resp => {
        var content = "";
        var found200 = false
        var amt = 0;
        resp.content.forEach((l) => {
            if (amt !== 15) {
                var logHTML = "";
                if (l.ResponseDetails.StatusCode == "200") {
                    found200 = true
                }

                logHTML += `
              <span class="${l.ResponseDetails.StatusCode == "200" ? "text-green-500" : "text-red-500"
                    }">[${l.ResponseDetails.StatusCode}]</span> <br>
              <span>Sent @ ${formatTime(Math.floor(new Date(l.ResponseDetails.SentAt).getTime() / 1000))}</span> <br> <span>Recv @ ${formatTime(Math.floor(new Date(l.ResponseDetails.RecvAt).getTime() / 1000))}<br></span>`;

                content += `<div class="bg-${contains(l.ResponseDetails.StatusCode) == "200" ? "green-600" : "red-600"} p-2 rounded-md shadow mt-4">
                <h1 class="text-md text-white font-mono">${l.Email}</h1>
                <h2 class="text-sm text-white font-mono">${"wip"}</h2>
        <div class="rounded-lg font-mono text-sm mt-2 p-3 bg-neutral ">
            <p>
              <p>
                ${logHTML}
              </p>
            </p>
        </div></div>`;
                amt++
            }
        })
        statusC = found200 ? "Yes" : "No";
        bgC = found200 ? "text-green-500" : "text-red-500";
        popInfo(`<h1 class="text-2xl">Logs for
        <span class="kbd">${resp.name}</span>
    </h1>`, `${content}`, function () { }, function () { });
    })
};

const add_task_to_db = (name, start, end, headurl) => {
    window.go.main.App.TaskAdd({ "name": name, "start": start, "end": end, "headurl": headurl })
    let info = document.getElementById("task_list").children;
    let found = false
    for (let i = 0; i < info.length; i++) {
        if (info[i].children[0].children[0].children[1].children[0].innerText == name) {
            found = true
        }
    }
    if (!found) {
        let html = get_task_html({
            name:name,
            headurl:headurl,
            start:start,
            end:end,
        });
        document.getElementById("task_list").innerHTML += html
    }
    
}

const delete_task_to_db = (name, start, end, headurl) => {
    if (window.go.main.App.TaskDelete({ "name": name, "start": start, "end": end, "headurl": headurl })) {
        let allNames = document.getElementById("task_list").children;
        for (let i = 0; i < allNames.length; i++) {
            let name = allNames[i].children[0].children[0].children[1].children[0].innerText.toLowerCase()
            if (allNames[i].children[0].children[0].children[1].children[0].innerText == name.toLowerCase()) {
                allNames[i].remove()
            }
        }
    }
}

const get_task_html = (task) => {
    if (task.headurl == undefined) {
        task.headurl = "https://s.namemc.com/2d/skin/face.png?id=44539e09e576d557&scale=4"
    }
    if (task.searches == undefined) {
        task.searches = "0"
    }
    if (!Number.isInteger(task.start)) {
        task.start = task.start_unix
    }
    if (!Number.isInteger(task.end)) {
        task.end = task.end_unix
    }

    id = new Date().getTime();
    return `
      <tr>
      <td>
                <div class="flex items-center gap-3">
                
                <div class="avatar">
                <label for="${id}">
                <div class=" shadow-lg rounded-lg w-10 btn-ghost transition transform hover:-translate-y-1 motion-reduce:transition-none motion-reduce:hover:transform-none">
                   <img src="${task.headurl}" class="rounded-lg">
                 </div>
               </label>
                <input type="checkbox" id="${id}" class="modal-toggle">
                <div class="modal items-center">
                <div class="modal-box flex absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2">
                    <div class="form-control">
                        <label class="label">
                        <div class=" shadow-lg rounded-lg w-100 btn-ghost transition transform hover:-translate-y-1 motion-reduce:transition-none motion-reduce:hover:transform-none">
                        <img src="${task.headurl.replace("&scale=4", "&scale=10")}" class="rounded-lg">
                      </div>
                        </label>
                        <div class="modal-action">

                            <button onclick="add_task_to_db('${task.name}',${task.start},${task.end},'${task.headurl}')" for="${id}" class="btn">
                                Add to queue
                            </button>
                            <button onclick="delete_task_to_db('${task.name}',${task.start},${task.end},'${task.headurl}')" for="${id}" class="btn">
                                Remove From Queue
                            </button>
                            <button onclick="add_logs('${task.name}')" for="${id}" class="btn">
                                Get Logs
                            </button>
                            <label for="${id}" class="btn">Close</label>
                        </div>
                    </div>
                </div>
            </div>
                </div>
                <div>
                  <div class="font-bold">${task.name}</div>
                  <div class="text-sm opacity-50">${task.searches}</div>
                </div>
                </div>
                </td> 
      <td>
          <span class="font-mono">
              ${formatTime(task.start) || task.start || "-"}
          </span>

      </td> 
      <td>
          <span class="font-mono">
              ${formatTime(task.end) || task.end || "-"}
          </span>
      </td> 
    </tr>`;
};

const add_task_html = (task) => {
    let html = get_task_html(task);
    document.getElementById("task_list").innerHTML += html;
};

const add_task_html_3c = (task) => {
    let html = get_task_html(task);
    document.getElementById("task_list_3c").innerHTML += html;
};

const check_if_logged_in = () => {
    if (getCookie("info_csrf") !== "") {
        window.location.replace('/home')
        return
    }
    window.location.replace('/discord/login')
}

const add_task_handler = () => {
    let name = document.getElementById("task_name");
    let value = name.value.trim().replace(/\t/g, "");
    name.value = "";
    window.go.main.App.GetRequestData(value).then(r => {
        let packet = JSON.parse(r);
        if (packet["code"] == 200) {
            if (packet.data.searches.headurl == "") {
                task.headurl = "https://s.namemc.com/2d/skin/face.png?id=a7a91811736bb2cb&scale=10"
            }
            if (packet.data.searches.status == "Locked" || packet.data.searches.status == "Possibly Available") {
                id = new Date().getTime();
                document.getElementById("task_list").innerHTML += `<tr>
                <td>
                <div class="flex items-center gap-3">
                
                <div class="avatar">
                <label for="${id}">
                <div class=" shadow-lg rounded-lg w-10 btn-ghost transition transform hover:-translate-y-1 motion-reduce:transition-none motion-reduce:hover:transform-none">
                   <img src="${packet.data.searches.headurl}" class="rounded-lg">
                 </div>
               </label>
                <input type="checkbox" id="${id}" class="modal-toggle">
                <div class="modal items-center">
                <div class="modal-box flex absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2">
                    <div class="form-control">
                        <label class="label">
                        <div class=" shadow-lg rounded-lg w-100 btn-ghost transition transform hover:-translate-y-1 motion-reduce:transition-none motion-reduce:hover:transform-none">
                        <img src="${packet.data.searches.headurl.replace("&scale=4", "&scale=10")}" class="rounded-lg">
                      </div>
                        </label>
                        <div class="modal-action">

                            <button onclick="add_task_to_db('${packet.data.Name}', ${packet.data.searches.start_unix}, ${packet.data.searches.end_unix}, '${packet.data.searches.headurl}')" for="${id}" class="btn">
                                Add to queue
                            </button>
                            <button onclick="delete_task_to_db('${packet.data.Name}', ${packet.data.searches.start_unix}, ${packet.data.searches.end_unix}, '${packet.data.searches.headurl}')" for="${id}" class="btn">
                                Remove From Queue
                            </button>
                            <button onclick="add_logs('${packet.data.Name}')" for="${id}" class="btn">
                                Get Logs
                            </button>
                            <label for="${id}" class="btn">Close</label>
                        </div>
                    </div>
                </div>
            </div>
                </div>
                <div>
                  <div class="font-bold">${packet.data.Name}</div>
                  <div class="text-sm opacity-50">${packet.data.searches.searches}</div>
                </div>
                </div>
                </td> 
                
                <td>
                ${formatTime(packet.data.searches.start_unix) || task.unix || "-"}
                </td>
                <td>
                ${formatTime(packet.data.searches.end_unix) || task.unix || "-"}
                </td>
                </tr>`;
            }
        }
    })
};