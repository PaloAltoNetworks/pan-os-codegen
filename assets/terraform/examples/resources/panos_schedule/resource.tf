resource "panos_schedule" "non_recurring" {
  location = {
    device_group = {
      name = panos_device_group.example.name
    }
  }

  name             = "non-recurring-schedule"
  disable_override = "yes"
  schedule_type = {
    non_recurring = [
      "2025/01/01@00:00-2025/01/31@23:59",
      "2025/12/24@00:00-2025/12/26@23:59"
    ]
  }
}

resource "panos_schedule" "daily" {
  location = {
    device_group = {
      name = panos_device_group.example.name
    }
  }

  name             = "daily-schedule"
  disable_override = "yes"
  schedule_type = {
    recurring = {
      daily = ["09:00-17:00", "18:00-22:00"]
    }
  }
}

resource "panos_schedule" "weekly" {
  location = {
    device_group = {
      name = panos_device_group.example.name
    }
  }

  name             = "weekly-schedule"
  disable_override = "yes"
  schedule_type = {
    recurring = {
      weekly = {
        monday    = ["08:00-12:00", "13:00-17:00"]
        tuesday   = ["08:00-12:00", "13:00-17:00"]
        wednesday = ["08:00-12:00", "13:00-17:00"]
        thursday  = ["08:00-12:00", "13:00-17:00"]
        friday    = ["08:00-12:00", "13:00-17:00"]
        saturday  = ["10:00-14:00"]
        sunday    = ["10:00-14:00"]
      }
    }
  }
}

resource "panos_device_group" "example" {
  location = {
    panorama = {}
  }

  name = "example-device-group"
}
